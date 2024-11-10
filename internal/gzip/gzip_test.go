package gzip

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockResponseWriter реализует mock для http.ResponseWriter
type MockResponseWriter struct {
	mock.Mock
	HeaderData http.Header
}

func (m *MockResponseWriter) Header() http.Header {
	return m.HeaderData
}

func (m *MockResponseWriter) Write(b []byte) (int, error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

func (m *MockResponseWriter) WriteHeader(statusCode int) {
	m.Called(statusCode)
}

// TestRequestDecompressor тестирует RequestDecompressor
func TestRequestDecompressor(t *testing.T) {
	// Создаем мок-объект для ResponseWriter
	mockWriter := new(MockResponseWriter)
	mockWriter.HeaderData = http.Header{}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что было установлено правильное содержимое
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, compressed world!"))
	})

	// Создаем gzip-запрос
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write([]byte("Hello, compressed world!"))
	gz.Close()

	req := httptest.NewRequest("POST", "/", &buf)
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Type", "text/html")

	// Установка ожидаемого результата
	mockWriter.On("WriteHeader", http.StatusOK)
	mockWriter.On("Write", mock.AnythingOfType("[]uint8")).Return(len([]byte("Hello, compressed world!")), nil)

	// Оборачиваем внутри RequestDecompressor
	decompressor := RequestDecompressor(handler)
	//recorder := httptest.NewRecorder()
	decompressor.ServeHTTP(mockWriter, req)

	// Проверка
	mockWriter.AssertExpectations(t)
}

// TestRequestDecompressor тестирует RequestDecompressor
func TestRequestDecompressor_WithAllowedContentTypes(t *testing.T) {
	// Создаем мок-объект для ResponseWriter
	mockWriter := new(MockResponseWriter)
	mockWriter.HeaderData = http.Header{}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что было установлено правильное содержимое
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, compressed world!"))
	})

	// Создаем gzip-запрос
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write([]byte("Hello, compressed world!"))
	gz.Close()

	req := httptest.NewRequest("POST", "/", &buf)
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "text/html")

	// Установка ожидаемого результата
	mockWriter.On("WriteHeader", http.StatusOK)
	mockWriter.On("Write", mock.AnythingOfType("[]uint8")).Return(len([]byte("Hello, compressed world!")), nil)

	// Оборачиваем внутри RequestDecompressor
	decompressor := RequestDecompressor(handler)
	//recorder := httptest.NewRecorder()
	decompressor.ServeHTTP(mockWriter, req)

	// Проверка
	mockWriter.AssertExpectations(t)
}

// TestRequestDecompressorWithGzip checks that the request is correctly decompressed
func TestRequestDecompressorWithGzip(t *testing.T) {
	// Create a gzip-compressed request body
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	zw.Write([]byte("Compressed content"))
	zw.Close()

	req := httptest.NewRequest("POST", "/", &buf)
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "text/html")

	// Capture the response
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err == nil {
			w.Write(body)
		}
	})

	decompressor := RequestDecompressor(handler)
	decompressor.ServeHTTP(recorder, req)

	// Check that the response is as expected
	responseBody := recorder.Body.String()
	assert.Equal(t, "Compressed content", responseBody)
	assert.Equal(t, http.StatusOK, recorder.Code)
}
