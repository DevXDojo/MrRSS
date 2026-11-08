package httputil

import (
	"compress/gzip"
	"mime"
	"net/http"
	"strconv"
	"strings"
)

// CompressStatic negotiates gzip for static text responses. Apply it to the
// frontend file handler only, not to API, media, or streaming handlers.
func CompressStatic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := &staticGzipWriter{
			ResponseWriter: w,
			canCompress: r.Method == http.MethodGet && r.Header.Get("Range") == "" &&
				!hasNoTransform(r.Header.Values("Cache-Control")) && acceptsStaticGzip(r.Header.Values("Accept-Encoding")),
		}
		defer writer.close()
		next.ServeHTTP(writer, r)
	})
}

type staticGzipWriter struct {
	http.ResponseWriter
	canCompress bool
	status      int
	header      http.Header
	committed   bool
	gzip        *gzip.Writer
}

func (w *staticGzipWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func (w *staticGzipWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	if status >= 100 && status < 200 {
		w.ResponseWriter.WriteHeader(status)
		return
	}
	w.status = status
	// Wait for the first bytes only when compression is possible, so an empty
	// response keeps its original encoding and Content-Length.
	if status != http.StatusOK || !w.canCompress {
		w.commit(nil)
	} else {
		// Freeze explicit headers as net/http does, even while the first body
		// bytes are deferred to distinguish an empty response.
		w.header = w.Header().Clone()
	}
}

func (w *staticGzipWriter) Write(data []byte) (int, error) {
	if len(data) == 0 && !w.committed {
		if w.status == 0 {
			w.WriteHeader(http.StatusOK)
		}
		return 0, nil
	}
	if !w.committed {
		w.commit(data)
	}
	if w.gzip != nil {
		return w.gzip.Write(data)
	}
	return w.ResponseWriter.Write(data)
}

func (w *staticGzipWriter) commit(first []byte) {
	if w.committed {
		return
	}
	w.committed = true
	if w.status == 0 {
		w.status = http.StatusOK
	}
	header := w.Header()
	if w.header != nil {
		clear(header)
		for key, values := range w.header {
			header[key] = values
		}
	}
	addStaticEncodingVary(header)
	if w.canCompress && w.status == http.StatusOK && len(first) > 0 &&
		header.Get("Content-Encoding") == "" && header.Get("Content-Range") == "" &&
		!hasNoTransform(header.Values("Cache-Control")) {
		if _, present := header["Content-Type"]; !present {
			header.Set("Content-Type", http.DetectContentType(first))
		}
		if isStaticTextType(header.Get("Content-Type")) {
			header.Del("Content-Length")
			header.Set("Content-Encoding", "gzip")
			w.gzip = gzip.NewWriter(w.ResponseWriter)
		}
	}
	w.ResponseWriter.WriteHeader(w.status)
}

// Flush commits an unstarted response without compression; active compressed
// responses flush both buffers. Unwrap also preserves ResponseController use.
func (w *staticGzipWriter) Flush() {
	if !w.committed {
		w.commit(nil)
	}
	if w.gzip != nil {
		_ = w.gzip.Flush()
	}
	_ = http.NewResponseController(w.ResponseWriter).Flush()
}

func (w *staticGzipWriter) close() error {
	if !w.committed {
		w.commit(nil)
	}
	if w.gzip != nil {
		return w.gzip.Close()
	}
	return nil
}

func isStaticTextType(value string) bool {
	mediaType, _, err := mime.ParseMediaType(value)
	if err != nil {
		return false
	}
	switch mediaType {
	case "text/html", "text/css", "text/javascript", "text/plain",
		"application/javascript", "application/x-javascript", "application/json", "image/svg+xml":
		return true
	}
	return false
}

func addStaticEncodingVary(header http.Header) {
	for _, value := range header.Values("Vary") {
		for _, field := range strings.Split(value, ",") {
			field = strings.TrimSpace(field)
			if field == "*" || strings.EqualFold(field, "Accept-Encoding") {
				return
			}
		}
	}
	header.Add("Vary", "Accept-Encoding")
}

func hasNoTransform(values []string) bool {
	for _, value := range values {
		for _, directive := range strings.Split(value, ",") {
			if strings.EqualFold(strings.TrimSpace(strings.SplitN(directive, "=", 2)[0]), "no-transform") {
				return true
			}
		}
	}
	return false
}

func acceptsStaticGzip(values []string) bool {
	gzipQuality, wildcardQuality, identityQuality := -1.0, -1.0, -1.0
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			parts := strings.Split(item, ";")
			coding := strings.ToLower(strings.TrimSpace(parts[0]))
			quality := 1.0
			for _, parameter := range parts[1:] {
				name, raw, found := strings.Cut(parameter, "=")
				if strings.EqualFold(strings.TrimSpace(name), "q") {
					parsed, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
					if !found || err != nil || !(parsed >= 0 && parsed <= 1) {
						quality = 0
						break
					}
					quality = parsed
				}
			}
			switch coding {
			case "gzip":
				gzipQuality = max(gzipQuality, quality)
			case "*":
				wildcardQuality = max(wildcardQuality, quality)
			case "identity":
				identityQuality = max(identityQuality, quality)
			}
		}
	}
	if gzipQuality < 0 {
		gzipQuality = wildcardQuality
	}
	return gzipQuality > 0 && gzipQuality >= identityQuality
}
