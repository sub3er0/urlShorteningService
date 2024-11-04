package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
	DatabaseDsn     string
}

var isParsed bool

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
