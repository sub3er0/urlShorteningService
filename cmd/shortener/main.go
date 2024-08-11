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
)

var shortenerInstance *shortener.URLShortener

func main() {
	cfg, err := config.InitConfig()

	if err != nil {
		log.Fatalf("Error while initializing config: %v", err)
	}

	pgStorage := storage.PgStorage{}

	if cfg.DatabaseDsn != "" {
		pgStorage.Init(cfg.DatabaseDsn)
		defer pgStorage.Close()
	}

	shortenerInstance = &shortener.URLShortener{
		Storage:       &storage.InMemoryStorage{Urls: make(map[string]string)},
		ServerAddress: cfg.ServerAddress,
		BaseURL:       cfg.BaseURL,
		DataStorage:   &storage.FileStorage{FileStoragePath: cfg.FileStoragePath},
		DBStorage:     pgStorage,
	}
	err = shortenerInstance.LoadData()

	if err != nil {
		log.Printf("In memmory storage fail: %v", err)
	}

	zapLogger, err := zap.NewDevelopment()

	if err != nil {
		log.Fatalf("Error creating logger: %s", err)
	}

	defer zapLogger.Sync()
	logger.Sugar = *zapLogger.Sugar()
	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Use(gzip.RequestDecompressor)
	r.Post("/", shortenerInstance.PostHandler)
	r.Get("/{id}", shortenerInstance.GetHandler)
	r.Get("/ping", shortenerInstance.PingHandler)
	r.Post("/api/shorten", shortenerInstance.JSONPostHandler)
	err = http.ListenAndServe(cfg.ServerAddress, r)

	if err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}
