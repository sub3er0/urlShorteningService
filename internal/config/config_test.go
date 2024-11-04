package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitConfig_ValidArgs(t *testing.T) {
	// Указываем аргументы командной строки
	os.Args = []string{"cmd", "-a", "localhost:8080", "-b", "http://localhost:8080/", "-f", "/path/to/file", "-d", "user:pass@/dbname"}

	// Act
	cfg, err := InitConfig()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:8080/", cfg.BaseURL)
	assert.Equal(t, "localhost:8080", cfg.ServerAddress)
	assert.Equal(t, "/path/to/file", cfg.FileStoragePath)
	assert.Equal(t, "user:pass@/dbname", cfg.DatabaseDsn)
}

//func TestInitConfig_ValidEnvVars(t *testing.T) {
//	// Установка переменных окружения
//	os.Setenv("SERVER_ADDRESS", "env.localhost:8080")
//	os.Setenv("BASE_URL", "http://env.localhost:8080/")
//	os.Setenv("FILE_STORAGE_PATH", "/env/path/to/file")
//	os.Setenv("DATABASE_DSN", "env:user:pass@/envdbname")
//
//	// Не забудьте сбросить переменные окружения после теста
//	defer os.Unsetenv("SERVER_ADDRESS")
//	defer os.Unsetenv("BASE_URL")
//	defer os.Unsetenv("FILE_STORAGE_PATH")
//	defer os.Unsetenv("DATABASE_DSN")
//
//	// Act
//	cfg, err := InitConfig()
//
//	// Assert
//	assert.NoError(t, err)
//	assert.Equal(t, "http://env.localhost:8080/", cfg.BaseURL)
//	assert.Equal(t, "env.localhost:8080", cfg.ServerAddress)
//	assert.Equal(t, "/env/path/to/file", cfg.FileStoragePath)
//	assert.Equal(t, "env:user:pass@/envdbname", cfg.DatabaseDsn)
//}
//
//func TestInitConfig_MissingServerAddress(t *testing.T) {
//	// Указываем аргументы командной строки, чтобы не передавать ServerAddress
//	os.Args = []string{"cmd", "-f", "/path/to/file", "-d", "user:pass@/dbname", "-a", ""}
//
//	// Act
//	cfg, err := InitConfig()
//
//	// Assert
//	assert.Nil(t, cfg)
//	assert.EqualError(t, err, "ServerAddress is required")
//}
//
//func TestInitConfig_MissingBaseURL(t *testing.T) {
//	// Указываем аргументы командной строки с отсутствующим BaseURL
//	os.Setenv("SERVER_ADDRESS", "env.localhost:8080")
//	defer os.Unsetenv("SERVER_ADDRESS")
//	os.Args = []string{"cmd", "-a", "localhost:8080", "-b", ""}
//
//	// Act
//	cfg, err := InitConfig()
//
//	// Assert
//	assert.Nil(t, cfg)
//	assert.EqualError(t, err, "BaseURL is required")
//}
