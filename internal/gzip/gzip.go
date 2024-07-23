package gzip

import (
	"compress/gzip"
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

func GzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") &&
			contains(r.Header.Get("Content-Type"), AllowedContentTypes) {
			w.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(w)
			reader, err := gzip.NewReader(r.Body)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			defer func(reader *gzip.Reader) {
				err := reader.Close()
				if err != nil {

				}
			}(reader)

			defer func(gz *gzip.Writer) {
				err := gz.Close()
				if err != nil {
					log.Fatalf("Error closing gzip.Writer: %s", err)
				}
			}(gz)

			rw := &gzipResponseWriter{w: w, gz: gz}
			r.Body = reader
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
