package gzip

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"
)

var AllowedContentTypes = []string{"application/json", "text/html"}

type gzipResponseWriter struct {
	w  http.ResponseWriter
	gz *gzip.Writer
}

func (rw *gzipResponseWriter) Header() http.Header {
	return rw.w.Header()
}

func (rw *gzipResponseWriter) Write(b []byte) (int, error) {
	return rw.gz.Write(b)
}

func (rw *gzipResponseWriter) WriteHeader(statusCode int) {
	rw.w.WriteHeader(statusCode)
}

func GzipResponseCompressor(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") &&
			contains(r.Header.Get("Content-Type"), AllowedContentTypes) {
			gz := gzip.NewWriter(w)

			defer func(gz *gzip.Writer) {
				err := gz.Close()
				if err != nil {
					log.Fatalf("Error closing gzip.Writer: %s", err)
				}
			}(gz)

			rw := &gzipResponseWriter{w: w, gz: gz}
			rw.Header().Set("Content-Encoding", "gzip")
			h.ServeHTTP(rw, r)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

func contains(target string, list []string) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}

	return false
}

func GzipRequestDecompressor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := w

		// Проверка на сжатие ответа
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gz := gzip.NewWriter(w)
			defer func(gz *gzip.Writer) {
				err := gz.Close()
				if err != nil {
					log.Fatalf("Error closing gzip.Writer: %s", err)
				}
			}(gz)

			rw = &gzipResponseWriter{w: w, gz: gz}
			rw.Header().Set("Content-Encoding", "gzip")
		}

		// Распаковка запроса
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			reader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(rw, "Error decompressing request body", http.StatusInternalServerError)
				return
			}
			defer reader.Close()
			r.Body = io.NopCloser(reader)
		}

		next.ServeHTTP(rw, r)
	})
}
