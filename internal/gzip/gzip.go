package gzip

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"
)

// AllowedContentTypes содержит список разрешенных типов контента для сжатия.
var AllowedContentTypes = []string{"application/json", "text/html"}

// gzipResponseWriter реализует интерфейс http.ResponseWriter и оборачивает gzip.Writer.
type gzipResponseWriter struct {
	w  http.ResponseWriter
	gz *gzip.Writer
}

// Header возвращает заголовки ответа HTTP.
func (rw *gzipResponseWriter) Header() http.Header {
	return rw.w.Header()
}

// Write записывает данные в сжатом формате.
func (rw *gzipResponseWriter) Write(b []byte) (int, error) {
	return rw.gz.Write(b)
}

// WriteHeader устанавливает код состояния для ответа.
func (rw *gzipResponseWriter) WriteHeader(statusCode int) {
	rw.w.WriteHeader(statusCode)
}

// RequestDecompressor возвращает обработчик, который распаковывает сжатые запросы и устанавливает сжатие для ответов.
func RequestDecompressor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := w

		// Проверка на сжатие ответа
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") &&
			contains(r.Header.Get("Content-Type"), AllowedContentTypes) {
			gz := gzip.NewWriter(w)
			defer func(gz *gzip.Writer) {
				err := gz.Close()

				if err != nil {
					log.Printf("Error closing gzip.Writer: %s", err)
					http.Error(rw, "Internal Server Error", http.StatusInternalServerError)

					return
				}
			}(gz)

			rw = &gzipResponseWriter{w: w, gz: gz}
			rw.Header().Set("Content-Encoding", "gzip")
		}

		// Распаковка запроса
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			reader, err := gzip.NewReader(r.Body)
			if err != nil {
				log.Printf("Error decompressing request body %s", err)
				http.Error(rw, "Error decompressing request body", http.StatusInternalServerError)

				return
			}
			defer reader.Close()
			r.Body = io.NopCloser(reader)
		}

		next.ServeHTTP(rw, r)
	})
}

// contains проверяет наличие строки в списке.
func contains(target string, list []string) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}

	return false
}
