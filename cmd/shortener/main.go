package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sub3er0/urlShorteningService/internal/config"
	"github.com/sub3er0/urlShorteningService/internal/shortener"
	"github.com/sub3er0/urlShorteningService/internal/storage"
	"go.uber.org/zap"
	"log"
	"net/http"
	"time"
)

var sugar zap.SugaredLogger

func main() {
	cfg, err := config.InitConfig()

	if err != nil {
		log.Fatalf("Error while initializing config: %v", err)
	}

	shortenerInstance := &shortener.URLShortener{
		Storage:       &storage.InMemoryStorage{Urls: make(map[string]string)},
		ServerAddress: cfg.ServerAddress,
		BaseURL:       cfg.BaseURL,
	}
	logger, err := zap.NewDevelopment()

	if err != nil {
		log.Fatalf("Error creating logger: %s", err)
	}

	defer logger.Sync()
	sugar = *logger.Sugar()
	r := chi.NewRouter()
	r.Use(responseLogger) // Log response details
	r.Use(requestLogger)  // Log request details
	r.Post("/", shortenerInstance.PostHandler)
	r.Get("/{id}", shortenerInstance.GetHandler)
	err = http.ListenAndServe(cfg.ServerAddress, r)

	if err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}

func responseLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		defer func() {
			log.Printf("Response: Status=%d, ContentLength=%d", rw.Status(), rw.BytesWritten())
		}()

		next.ServeHTTP(rw, r)
	})
}

func requestLogger(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method
		h.ServeHTTP(w, r)
		duration := time.Since(start)
		sugar.Infoln(
			"uri", uri,
			"method", method,
			"duration", duration,
		)
	}

	return http.HandlerFunc(logFn)
}
