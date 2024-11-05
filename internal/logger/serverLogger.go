package logger

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// Sugar является экземпляром логгера с уровнями информации и отладки.
var Sugar zap.SugaredLogger

// RequestLogger создает HTTP-обработчик, который логирует информацию о запросах и
// ответах, используя логгер Sugar.
func RequestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method
		rw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			Sugar.Infoln(
				fmt.Sprintf("Response: Status=%d", rw.Status()),
				fmt.Sprintf("ContentLength=%d", rw.BytesWritten()),
			)

			duration := time.Since(start)
			Sugar.Infoln(
				"uri", uri,
				"method", method,
				"duration", duration,
			)
		}()

		h.ServeHTTP(rw, r)
	})
}
