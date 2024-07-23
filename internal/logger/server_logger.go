package logger

import (
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"log"
	"net/http"
	"time"
)

var Sugar zap.SugaredLogger

func ResponseLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		defer func() {
			log.Printf("Response: Status=%d, ContentLength=%d", rw.Status(), rw.BytesWritten())
		}()

		next.ServeHTTP(rw, r)
	})
}

func RequestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method
		h.ServeHTTP(w, r)
		duration := time.Since(start)
		Sugar.Infoln(
			"uri", uri,
			"method", method,
			"duration", duration,
		)
	})
}
