package logger

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// Функция для инициализации логгера перед тестами
func initLogger() {
	logger, _ := zap.NewDevelopment()
	Sugar = *logger.Sugar()
}

func TestRequestLogger(t *testing.T) {
	initLogger() // Инициализация логгера

	// Создаем тестовый обработчик
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)     // Записываем код состояния
		w.Write([]byte("Hello, World!")) // Записываем ответ
	})

	// Оборачиваем наш обработчик в RequestLogger
	loggedHandler := RequestLogger(handler)

	// Создаем тестовый HTTP-запрос
	req := httptest.NewRequest("GET", "/test", nil)

	// Рекордер для захвата ответа
	recorder := httptest.NewRecorder()

	// Выполнить запрос
	loggedHandler.ServeHTTP(recorder, req)

	// Проверяем статус код
	assert.Equal(t, http.StatusOK, recorder.Code)

	// Проверяем, что тело ответа соответствует ожидаемому
	assert.Equal(t, "Hello, World!", recorder.Body.String())
}
