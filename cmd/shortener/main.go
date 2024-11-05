package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sub3er0/urlShorteningService/internal/config"
	"github.com/sub3er0/urlShorteningService/internal/cookie"
	"github.com/sub3er0/urlShorteningService/internal/gzip"
	"github.com/sub3er0/urlShorteningService/internal/logger"
	"github.com/sub3er0/urlShorteningService/internal/repository"
	"github.com/sub3er0/urlShorteningService/internal/shortener"
	"github.com/sub3er0/urlShorteningService/internal/storage"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var shortenerInstance *shortener.URLShortener

func main() {
	cfg, err := config.InitConfig()

	if err != nil {
		log.Fatalf("Error while initializing config: %v", err)
	}

	var dataUrlsStorage storage.URLStorageInterface
	var dataUsersStorage storage.UserStorageInterface

	if cfg.DatabaseDsn != "" {
		dsn := cfg.DatabaseDsn
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

		if err != nil {
			log.Fatalf("Failed to connect database: %v", err)
		}

		err = db.AutoMigrate(&storage.URL{})

		if err != nil {
			log.Fatalf("Failed to connect database: %v", err)
		}

		err = db.AutoMigrate(&storage.UserCookie{})

		if err != nil {
			log.Fatalf("Failed to connect database: %v", err)
		}

		dataUrlsStorage = &storage.URLStorage{}
		dataUrlsStorage.Init(cfg.DatabaseDsn)
		defer dataUrlsStorage.Close()

		dataUsersStorage = &storage.UsersStorage{}
		dataUsersStorage.Init(cfg.DatabaseDsn)
		defer dataUsersStorage.Close()
	} else if cfg.FileStoragePath != "" {
		dataUrlsStorage = &storage.FileStorage{FileStoragePath: cfg.FileStoragePath}
		dataUsersStorage = &storage.FileStorage{FileStoragePath: cfg.FileStoragePath}
	} else {
		dataUrlsStorage = &storage.InMemoryStorage{Urls: make(map[string]string)}
		dataUsersStorage = &storage.InMemoryStorage{Urls: make(map[string]string)}
	}

	cookieManager := cookie.CookieManager{
		Storage: dataUsersStorage,
	}

	var userRepository repository.UserRepositoryInterface
	var urlRepository repository.UrlRepositoryInterface

	urlRepository = &repository.UrlRepository{Storage: dataUrlsStorage}
	userRepository = &repository.UserRepository{Storage: dataUsersStorage}

	shortenerInstance = &shortener.URLShortener{
		UserRepository: userRepository,
		UrlRepository:  urlRepository,
		ServerAddress:  cfg.ServerAddress,
		BaseURL:        cfg.BaseURL,
		CookieManager:  &cookieManager,
		RemoveChan:     make(chan string),
	}

	go shortenerInstance.Worker()

	zapLogger, err := zap.NewDevelopment()

	if err != nil {
		log.Fatalf("Error creating logger: %s", err)
	}

	defer zapLogger.Sync()
	logger.Sugar = *zapLogger.Sugar()
	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Use(gzip.RequestDecompressor)
	r.With(cookieManager.CookieHandler).Route("/", func(r chi.Router) {
		r.Post("/", shortenerInstance.PostHandler)
		r.Get("/{id}", shortenerInstance.GetHandler)
		r.Post("/api/shorten", shortenerInstance.JSONPostHandler)
		r.Post("/api/shorten/batch", shortenerInstance.JSONBatchHandler)

		r.With(cookieManager.AuthMiddleware).Get("/api/user/urls", shortenerInstance.GetUserUrls)
		r.With(cookieManager.AuthMiddleware).Delete("/api/user/urls", shortenerInstance.DeleteUserUrls)
	})

	r.Get("/ping", shortenerInstance.PingHandler)

	err = http.ListenAndServe(cfg.ServerAddress, r)

	if err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}
