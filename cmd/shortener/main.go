package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/sub3er0/urlShorteningService/internal/config"
	"github.com/sub3er0/urlShorteningService/internal/gzip"
	"github.com/sub3er0/urlShorteningService/internal/logger"
	"github.com/sub3er0/urlShorteningService/internal/shortener"
	"github.com/sub3er0/urlShorteningService/internal/storage"
	"go.uber.org/zap"
	"log"
	"net/http"
	"strings"
)

var shortenerInstance *shortener.URLShortener

func main() {
	cfg, err := config.InitConfig()

	if err != nil {
		log.Fatalf("Error while initializing config: %v", err)
	}

	shortenerInstance = &shortener.URLShortener{
		Storage:       &storage.InMemoryStorage{Urls: make(map[string]string)},
		ServerAddress: cfg.ServerAddress,
		BaseURL:       cfg.BaseURL,
	}
	zapLogger, err := zap.NewDevelopment()

	if err != nil {
		log.Fatalf("Error creating logger: %s", err)
	}

	defer zapLogger.Sync()
	logger.Sugar = *zapLogger.Sugar()
	r := chi.NewRouter()
	r.Use(logger.ResponseLogger)
	r.Use(logger.RequestLogger)
	r.Use(gzip.GzipMiddleware)
	r.Use(jsonMiddleware)
	r.Post("/", shortenerInstance.PostHandler)
	r.Get("/{id}", shortenerInstance.GetHandler)
	r.Post("/api/shorten", shortenerInstance.JSONPostHandler)
	err = http.ListenAndServe(cfg.ServerAddress, r)

	if err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}

func jsonMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Type"), "application/json") &&
			r.Method == http.MethodPost {
			h = http.HandlerFunc(shortenerInstance.JSONPostHandler)
			h.ServeHTTP(w, r)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}
