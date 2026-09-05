package update

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func downloadUpdateFile(ctx context.Context, client *http.Client, downloadURL, filePath, requestID string, maxAttempts int, baseDelay time.Duration) (int64, int64, error) {
	partPath := filePath + ".part"
	_ = os.Remove(partPath)
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		written, total, retry, err := downloadUpdateAttempt(ctx, client, downloadURL, partPath, requestID)
		if err == nil {
			if err := os.Remove(filePath); err != nil && !errors.Is(err, os.ErrNotExist) {
				_ = os.Remove(partPath)
				return 0, 0, fmt.Errorf("replace existing update: %w", err)
			}
			if err := os.Rename(partPath, filePath); err != nil {
				_ = os.Remove(partPath)
				return 0, 0, fmt.Errorf("finalize update: %w", err)
			}
			return written, total, nil
		}
		lastErr = err
		if !retry || attempt == maxAttempts || ctx.Err() != nil {
			break
		}
		delay := baseDelay * time.Duration(1<<(attempt-1))
		if delay > 0 {
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				lastErr = ctx.Err()
				attempt = maxAttempts
			case <-timer.C:
			}
		}
	}

	_ = os.Remove(partPath)
	return 0, 0, lastErr
}

func downloadUpdateAttempt(ctx context.Context, client *http.Client, downloadURL, partPath, requestID string) (int64, int64, bool, error) {
	offset := int64(0)
	if info, err := os.Stat(partPath); err == nil {
		offset = info.Size()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return 0, 0, false, err
	}
	req.Header.Set("User-Agent", "MrRSS-Updater")
	req.Header.Set("Accept", "application/octet-stream")
	if offset > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", offset))
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, isRetryableDownloadError(err), err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusRequestedRangeNotSatisfiable && offset > 0 {
		_ = os.Remove(partPath)
		return 0, 0, true, fmt.Errorf("range no longer valid")
	}
	if resp.StatusCode >= 500 {
		return 0, 0, true, fmt.Errorf("download server returned %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return 0, 0, false, fmt.Errorf("download server returned %d", resp.StatusCode)
	}

	resume := offset > 0 && resp.StatusCode == http.StatusPartialContent
	if resp.StatusCode == http.StatusPartialContent {
		start, _, ok := parseContentRange(resp.Header.Get("Content-Range"))
		if !ok || start != offset {
			_ = os.Remove(partPath)
			return 0, 0, true, fmt.Errorf("invalid range response")
		}
	} else {
		offset = 0
	}

	flags := os.O_CREATE | os.O_WRONLY
	if resume {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}
	out, err := os.OpenFile(partPath, flags, 0o600)
	if err != nil {
		return 0, 0, false, err
	}

	total := responseTotalSize(resp, offset)
	setDownloadProgress(downloadProgress{
		RequestID: requestID, State: "downloading", BytesWritten: offset,
		TotalBytes: total, Percentage: downloadPercentage(offset, total), Indeterminate: total <= 0,
	})

	buffer := make([]byte, 64*1024)
	written := offset
	for {
		n, readErr := resp.Body.Read(buffer)
		if n > 0 {
			wn, writeErr := out.Write(buffer[:n])
			written += int64(wn)
			setDownloadProgress(downloadProgress{
				RequestID: requestID, State: "downloading", BytesWritten: written,
				TotalBytes: total, Percentage: downloadPercentage(written, total), Indeterminate: total <= 0,
			})
			if writeErr != nil {
				_ = out.Close()
				return written, total, false, writeErr
			}
			if wn != n {
				_ = out.Close()
				return written, total, false, io.ErrShortWrite
			}
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			_ = out.Close()
			return written, total, isRetryableDownloadError(readErr), readErr
		}
	}
	if err := out.Sync(); err != nil {
		_ = out.Close()
		return written, total, false, err
	}
	if err := out.Close(); err != nil {
		return written, total, false, err
	}
	if total > 0 && written != total {
		return written, total, true, fmt.Errorf("download incomplete: expected %d bytes, got %d", total, written)
	}
	return written, total, false, nil
}

func responseTotalSize(resp *http.Response, offset int64) int64 {
	if resp.StatusCode == http.StatusPartialContent {
		_, total, ok := parseContentRange(resp.Header.Get("Content-Range"))
		if ok {
			return total
		}
	}
	if resp.ContentLength > 0 {
		return offset + resp.ContentLength
	}
	return 0
}

func parseContentRange(value string) (int64, int64, bool) {
	if !strings.HasPrefix(value, "bytes ") {
		return 0, 0, false
	}
	parts := strings.Split(strings.TrimPrefix(value, "bytes "), "/")
	if len(parts) != 2 {
		return 0, 0, false
	}
	rangeParts := strings.Split(parts[0], "-")
	if len(rangeParts) != 2 {
		return 0, 0, false
	}
	start, err1 := strconv.ParseInt(rangeParts[0], 10, 64)
	total, err2 := strconv.ParseInt(parts[1], 10, 64)
	end, err3 := strconv.ParseInt(rangeParts[1], 10, 64)
	return start, total, err1 == nil && err2 == nil && err3 == nil && start >= 0 && end >= start && total > end
}

func isRetryableDownloadError(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	return errors.As(err, &netErr) || errors.Is(err, io.ErrUnexpectedEOF)
}

func classifyDownloadError(err error) string {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, os.ErrDeadlineExceeded) {
		return "download_timeout"
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return "download_timeout"
		}
		return "download_network_error"
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "server returned 5") {
		return "download_server_error"
	}
	if strings.Contains(message, "incomplete") {
		return "download_incomplete"
	}
	return "download_failed"
}

func downloadPercentage(written, total int64) float64 {
	if total <= 0 {
		return 0
	}
	percentage := float64(written) * 100 / float64(total)
	if percentage > 100 {
		return 100
	}
	return percentage
}
