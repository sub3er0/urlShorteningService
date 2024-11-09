package config

import (
	"flag"
	"fmt"
	"os"
)

// Config представляет конфигурацию приложения.
// Это структура содержит параметры, необходимые для настройки серверного приложения.
type Config struct {
	// ServerAddress определяет адрес HTTP-сервера, на котором будет работать приложение.
	ServerAddress string `json:"server_address"`

	// BaseURL представляет базовый адрес, который используется для сокращенных URL.
	BaseURL string `json:"base_url"`

	// FileStoragePath задает путь к файлу, где могут храниться данные.
	FileStoragePath string `json:"file_storage_path"`

	// DatabaseDsn представляет строку подключения к базе данных.
	DatabaseDsn string `json:"database_dsn"`
}

// isParsed отслеживает, выполнена ли обработка аргументов командной строки.
var isParsed bool

// InitConfig инициализирует конфигурацию приложения.
func InitConfig() (*Config, error) {
	cfg := &Config{}

	if !isParsed {
		flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080/", "Базовый адрес для сокращенных URL")
		flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "Адрес HTTP-сервера")
		flag.StringVar(&cfg.FileStoragePath, "f", "", "Путь до файла")
		flag.StringVar(
			&cfg.DatabaseDsn,
			"d", "",
			"Строка подключения к базе данных")

		flag.Parse()
		isParsed = true
	}

	if ServerAddress := os.Getenv("SERVER_ADDRESS"); ServerAddress != "" {
		cfg.ServerAddress = ServerAddress
	}

	if BaseURL := os.Getenv("BASE_URL"); BaseURL != "" {
		cfg.BaseURL = BaseURL
	}

	if FileStoragePath := os.Getenv("FILE_STORAGE_PATH"); FileStoragePath != "" {
		cfg.FileStoragePath = FileStoragePath
	}

	if DatabaseDsn := os.Getenv("DATABASE_DSN"); DatabaseDsn != "" {
		cfg.DatabaseDsn = DatabaseDsn
	}

	if cfg.ServerAddress == "" {
		return nil, fmt.Errorf("ServerAddress is required")
	}

	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("BaseURL is required")
	}

	return cfg, nil
}
