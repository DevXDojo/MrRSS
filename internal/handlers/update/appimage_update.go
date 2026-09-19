package update

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func validateUpdateDownload(path string) (string, error) {
	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("update path must be absolute")
	}
	path = filepath.Clean(path)
	dir := filepath.Dir(path)
	if filepath.Clean(filepath.Dir(dir)) != filepath.Clean(os.TempDir()) || !strings.HasPrefix(filepath.Base(dir), "mrrss-update-") {
		return "", fmt.Errorf("update must be in its download directory")
	}
	for _, entry := range []string{dir, path} {
		info, err := os.Lstat(entry)
		if err != nil || info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("invalid update download path")
		}
	}
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return "", fmt.Errorf("update must be a regular file")
	}
	return path, nil
}

func cleanupUpdateDownload(path string) {
	// Never recursively delete a caller-selected directory.
	if _, err := validateUpdateDownload(path); err != nil {
		return
	}
	if err := os.Remove(path); err != nil {
		log.Printf("Failed to remove update download: %v", err)
		return
	}
	if err := os.Remove(filepath.Dir(path)); err != nil {
		log.Printf("Failed to remove empty update directory: %v", err)
	}
}

// replaceAppImage stages beside the target, so rename is atomic even when the
// download directory is on another filesystem. The running mount retains its
// old inode until shutdown; the desktop launcher keeps pointing at the new one.
func replaceAppImage(ctx context.Context, download, installed string) (string, error) {
	if !filepath.IsAbs(installed) || !strings.EqualFold(filepath.Ext(download), ".appimage") {
		return "", fmt.Errorf("invalid AppImage path")
	}
	target, err := filepath.EvalSymlinks(installed)
	if err != nil {
		return "", fmt.Errorf("resolve installed AppImage: %w", err)
	}
	info, err := os.Stat(target)
	if err != nil || !info.Mode().IsRegular() {
		return "", fmt.Errorf("installed AppImage is not a regular file")
	}
	source, err := os.Open(download)
	if err != nil {
		return "", err
	}
	defer source.Close()
	// Reject HTML/error pages masquerading as a downloaded AppImage.
	header := make([]byte, 11)
	if _, err := io.ReadFull(source, header); err != nil || string(header[:4]) != "\x7fELF" || string(header[8:11]) != "AI\x02" {
		return "", fmt.Errorf("download is not a type 2 AppImage")
	}
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	stage, err := os.CreateTemp(filepath.Dir(target), ".mrrss-update-*")
	if err != nil {
		return "", fmt.Errorf("stage installed AppImage: %w", err)
	}
	defer os.Remove(stage.Name())
	defer stage.Close()
	if _, err := io.Copy(stage, &updateContextReader{ctx: ctx, reader: source}); err != nil {
		return "", err
	}
	if err := stage.Chmod(info.Mode().Perm() | 0700); err != nil {
		return "", err
	}
	if err := stage.Sync(); err != nil {
		return "", err
	}
	if err := stage.Close(); err != nil {
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if err := os.Rename(stage.Name(), target); err != nil {
		return "", fmt.Errorf("replace installed AppImage: %w", err)
	}
	return target, nil
}

type updateContextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r *updateContextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}
