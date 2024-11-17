package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/sub3er0/urlShorteningService/internal/config"
	"github.com/sub3er0/urlShorteningService/internal/cookie"
	"github.com/sub3er0/urlShorteningService/internal/gzip"
	"github.com/sub3er0/urlShorteningService/internal/logger"
	"github.com/sub3er0/urlShorteningService/internal/repository"
	"github.com/sub3er0/urlShorteningService/internal/shortener"
	"github.com/sub3er0/urlShorteningService/internal/storage"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net/http"
	"os"
)

var shortenerInstance *shortener.URLShortener

var (
	buildVersion string = "N/A" // Значение по умолчанию
	buildDate    string = "N/A" // Значение по умолчанию
	buildCommit  string = "N/A" // Значение по умолчанию
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	configuration := &config.Configuration{}

	cfg, err := configuration.InitConfig()

	if err != nil {
		log.Fatalf("Error while initializing configuration: %v", err)
	}

	var dataUrlsStorage storage.URLStorageInterface
	var dataUsersStorage storage.UserStorageInterface

	if cfg.DatabaseDsn != "" {
		defaultStorage := &storage.DefaultStorage{}
		defaultStorage.Init(cfg.DatabaseDsn)
		defer defaultStorage.Close()

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

	var urlRepository = &repository.URLRepository{Storage: dataUrlsStorage}
	var userRepository = &repository.UserRepository{Storage: dataUsersStorage}

	shortenerInstance = &shortener.URLShortener{
		UserRepository: userRepository,
		URLRepository:  urlRepository,
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

	// Проверка переменной окружения или флага для включения HTTPS
	if cfg.EnableHTTPS || os.Getenv("ENABLE_HTTPS") == "true" {
		// Создаем менеджер для автоматического управления сертификатами
		manager := &autocert.Manager{
			Cache:      autocert.DirCache("cache-dir"),      // Директория для хранения сертификатов
			Prompt:     autocert.AcceptTOS,                  // Принять условия использования
			HostPolicy: autocert.HostWhitelist("localhost"), // Перечень доменов
		}

		// Создаем сервер с поддержкой TLS
		server := &http.Server{
			Addr:      ":443",              // Порт для HTTPS
			Handler:   r,                   // Ваше обработчик
			TLSConfig: manager.TLSConfig(), // TLS-конфигурация
		}

		// Запускаем сервер
		log.Println("Starting server on port 443")
		if err := server.ListenAndServeTLS("", ""); err != nil {
			log.Fatalf("Error starting HTTPS server: %s", err)
		}
	} else {
		err = http.ListenAndServe(cfg.ServerAddress, r)
		if err != nil {
			log.Fatalf("Error starting server: %s", err)
		}
	}

	if err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}
