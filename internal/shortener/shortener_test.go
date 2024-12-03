package shortener

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/sub3er0/urlShorteningService/internal/repository"
	"github.com/sub3er0/urlShorteningService/internal/storage"
)

// MockURLRepository - мок для URLRepositoryInterface.
type MockURLRepository struct {
	mock.Mock
}

// GetURL - реализует метод интерфейса URLRepositoryInterface.
func (m *MockURLRepository) GetURL(shortURL string) (storage.GetURLRow, bool) {
	args := m.Called(shortURL)
	return args.Get(0).(storage.GetURLRow), args.Bool(1)
}

// GetURLCount - реализует метод интерфейса URLRepositoryInterface.
func (m *MockURLRepository) GetURLCount() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

// GetShortURL - реализует метод интерфейса URLRepositoryInterface.
func (m *MockURLRepository) GetShortURL(URL string) (string, error) {
	args := m.Called(URL)
	return args.String(0), args.Error(1)
}

// Save - реализует метод интерфейса URLRepositoryInterface.
func (m *MockURLRepository) Save(ShortURL string, URL string, userID string) error {
	args := m.Called(ShortURL, URL, userID)
	return args.Error(0)
}

// LoadData - реализует метод интерфейса URLRepositoryInterface.
func (m *MockURLRepository) LoadData() ([]storage.DataStorageRow, error) {
	args := m.Called()
	return args.Get(0).([]storage.DataStorageRow), args.Error(1)
}

// Ping - реализует метод интерфейса URLRepositoryInterface.
func (m *MockURLRepository) Ping() bool {
	args := m.Called()
	return args.Bool(0)
}

// SaveBatch - реализует метод интерфейса URLRepositoryInterface.
func (m *MockURLRepository) SaveBatch(dataStorageRows []storage.DataStorageRow) error {
	args := m.Called(dataStorageRows)
	return args.Error(0)
}

// MockUserRepository - мок для UserRepositoryInterface.
type MockUserRepository struct {
	mock.Mock
}

// IsUserExist - реализует метод интерфейса UserRepositoryInterface.
func (m *MockUserRepository) IsUserExist(uniqueID string) bool {
	args := m.Called(uniqueID)
	return args.Bool(0)
}

// SaveUser - реализует метод интерфейса UserRepositoryInterface.
func (m *MockUserRepository) SaveUser(uniqueID string) error {
	args := m.Called(uniqueID)
	return args.Error(0)
}

// GetUserUrls - реализует метод интерфейса UserRepositoryInterface.
func (m *MockUserRepository) GetUserUrls(uniqueID string) ([]storage.UserUrlsResponseBodyItem, error) {
	args := m.Called(uniqueID)
	return args.Get(0).([]storage.UserUrlsResponseBodyItem), args.Error(1)
}

// GetUsersCount получение количества пользователей
func (m *MockUserRepository) GetUsersCount() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

// DeleteUserUrls - реализует метод интерфейса UserRepositoryInterface.
func (m *MockUserRepository) DeleteUserUrls(uniqueID string, shortURLs []string) error {
	args := m.Called(uniqueID, shortURLs)
	return args.Error(0)
}

// MockCookieManager - мок для CookieManagerInterface.
type MockCookieManager struct {
	mock.Mock
	ActualCookieValue string
}

// CookieHandler - реализует метод CookieHandler интерфейса CookieManagerInterface.
func (m *MockCookieManager) CookieHandler(h http.Handler) http.Handler {
	args := m.Called(h)
	return args.Get(0).(http.Handler)
}

// AuthMiddleware - реализует метод AuthMiddleware интерфейса CookieManagerInterface.
func (m *MockCookieManager) AuthMiddleware(h http.Handler) http.Handler {
	args := m.Called(h)
	return args.Get(0).(http.Handler)
}

// GetActualCookieValue - реализует метод GetActualCookieValue интерфейса CookieManagerInterface.
func (m *MockCookieManager) GetActualCookieValue() string {
	args := m.Called()
	return args.String(0)
}

func TestWorker_SuccessfulDeletion(t *testing.T) {
	userRepo := new(MockUserRepository)
	cookieManager := &MockCookieManager{ActualCookieValue: "test_user_id"}

	us := &URLShortener{
		UserRepository: userRepo,
		CookieManager:  cookieManager,
		RemoveChan:     make(chan string, 2), // создаем буферизированный канал
	}

	shortURL := "shortURL1"
	userRepo.On("DeleteUserUrls", cookieManager.ActualCookieValue, []string{shortURL}).Return(nil)
	cookieManager.On("GetActualCookieValue").Return(cookieManager.ActualCookieValue)

	// Запускаем Worker в горутине
	go us.Worker()

	// Отправляем короткий URL в RemoveChan.
	us.RemoveChan <- shortURL

	// Закрываем RemoveChan, чтобы сигнализировать о завершении.
	close(us.RemoveChan)

	// Используем небольшую задержку, чтобы дать время для завершения будущих вызовов
	time.Sleep(100 * time.Millisecond) // Задержка для гарантии выполнения Worker

	// Проверяем, что метод DeleteUserUrls был вызван
	userRepo.AssertExpectations(t)
}

func BenchmarkWorker(b *testing.B) {
	userRepo := new(MockUserRepository)
	cookieManager := &MockCookieManager{ActualCookieValue: "test_user_id"}

	us := &URLShortener{
		UserRepository: userRepo,
		CookieManager:  cookieManager,
		RemoveChan:     make(chan string, 100),
	}

	for i := 0; i < b.N; i++ {
		shortURL := "shortURL" + strconv.Itoa(i) // Генерация тестового короткого URL
		userRepo.On("DeleteUserUrls", cookieManager.ActualCookieValue, []string{shortURL}).Return(nil)
		cookieManager.On("GetActualCookieValue").Return(cookieManager.ActualCookieValue)

		go us.Worker()

		us.RemoveChan <- shortURL
	}

	close(us.RemoveChan)

	time.Sleep(100 * time.Millisecond)
}

func TestWorker_ErrorDuringDeletion(t *testing.T) {
	userRepo := new(MockUserRepository)
	cookieManager := &MockCookieManager{ActualCookieValue: "test_user_id"}

	us := &URLShortener{
		UserRepository: userRepo,
		CookieManager:  cookieManager,
		RemoveChan:     make(chan string, 1),
	}

	shortURL := "shortURL1"
	cookieManager.On("GetActualCookieValue").Return(cookieManager.ActualCookieValue)
	userRepo.On("DeleteUserUrls", cookieManager.ActualCookieValue, []string{shortURL}).Return(errors.New("deletion error"))

	go us.Worker()

	us.RemoveChan <- shortURL
	close(us.RemoveChan)

	time.Sleep(100 * time.Millisecond) // Задержка для гарантии выполнения Worker

	userRepo.AssertExpectations(t) // Проверяем, что DeleteUserUrls был вызван, даже если произошла ошибка
}

func TestGetHandler_URLNotFound(t *testing.T) {
	mockRepo := new(MockURLRepository)
	us := &URLShortener{URLRepository: mockRepo}

	req := httptest.NewRequest("GET", "/url/unknownID", nil)
	w := httptest.NewRecorder()

	mockRepo.On("GetURL", "").Return(storage.GetURLRow{}, false) // Установка ожидания для неопознанного URL

	us.GetHandler(w, req)

	res := w.Result()
	res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
	mockRepo.AssertExpectations(t)
}

func TestGetHandler_URLExists(t *testing.T) {
	mockRepo := new(MockURLRepository)
	us := &URLShortener{URLRepository: mockRepo}

	req := httptest.NewRequest("GET", "/url/knownID", nil)
	w := httptest.NewRecorder()

	expectedURL := "http://example.com"
	storedRow := storage.GetURLRow{URL: expectedURL, IsDeleted: false}
	mockRepo.On("GetURL", "").Return(storedRow, true) // Установка ожидания для существующего URL

	// Act
	us.GetHandler(w, req)

	// Assert
	res := w.Result()
	res.Body.Close()
	assert.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)
	assert.Equal(t, expectedURL, w.Header().Get("Location"))
	mockRepo.AssertExpectations(t)
}

func TestGetHandler_URLIsDeleted(t *testing.T) {
	mockRepo := new(MockURLRepository)
	us := &URLShortener{URLRepository: mockRepo}

	req := httptest.NewRequest("GET", "/url/deletedID", nil)
	w := httptest.NewRecorder()

	storedRow := storage.GetURLRow{URL: "http://example.com", IsDeleted: true}
	mockRepo.On("GetURL", "").Return(storedRow, true) // Установка ожидания для удаленного URL

	us.GetHandler(w, req)

	res := w.Result()
	res.Body.Close()
	assert.Equal(t, http.StatusGone, res.StatusCode)
	mockRepo.AssertExpectations(t)
}

func TestPingHandler_SuccessfulConnection(t *testing.T) {
	mockRepo := new(MockURLRepository)
	us := &URLShortener{URLRepository: mockRepo}

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	mockRepo.On("Ping").Return(true) // Установка ожидания на успешный пинг

	us.PingHandler(w, req)

	res := w.Result()
	res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	mockRepo.AssertExpectations(t)
}

func TestPingHandler_ConnectionError(t *testing.T) {
	mockRepo := new(MockURLRepository)
	us := &URLShortener{URLRepository: mockRepo}

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	mockRepo.On("Ping").Return(false) // Установка ожидания на ошибку пинга

	us.PingHandler(w, req)

	res := w.Result()
	res.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	mockRepo.AssertExpectations(t)
}

// MockURLShortener - мок для интерфейса URLShortenerInterface
type MockURLShortener struct {
	mock.Mock
	URLRepository repository.URLRepositoryInterface
	BaseURL       string
}

// GetHandler - реализует метод интерфейса
func (m *MockURLShortener) GetHandler(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
}

// PingHandler - реализует метод интерфейса
func (m *MockURLShortener) PingHandler(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
}

// JSONPostHandler - реализует метод интерфейса
func (m *MockURLShortener) JSONPostHandler(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
}

// JSONBatchHandler - реализует метод интерфейса
func (m *MockURLShortener) JSONBatchHandler(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
}

// GetUserUrls - реализует метод интерфейса
func (m *MockURLShortener) GetUserUrls(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
}

// DeleteUserUrls - реализует метод интерфейса
func (m *MockURLShortener) DeleteUserUrls(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
}

// Worker - реализует метод интерфейса
func (m *MockURLShortener) Worker() {
	m.Called()
}

func TestJSONPostHandler_Success(t *testing.T) {
	mockRepo := new(MockURLRepository)
	mockCookieManager := new(MockCookieManager)
	us := &URLShortener{
		URLRepository: mockRepo,
		CookieManager: mockCookieManager,
		BaseURL:       "http://short.url/",
	}

	requestBody := RequestBody{URL: "http://example.com"}
	jsonBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/shorten", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Установка ожидания
	mockRepo.On("GetShortURL", mock.Anything).Return("", errors.New("short url not found"))
	mockRepo.On("Save", mock.Anything, requestBody.URL, mock.Anything).Return(nil)
	mockCookieManager.On("GetActualCookieValue").Return("")

	// Act
	us.JSONPostHandler(w, req)

	// Assert
	res := w.Result()
	res.Body.Close()
	assert.Equal(t, http.StatusCreated, res.StatusCode)

	var responseBody JSONResponseBody
	_ = json.NewDecoder(res.Body).Decode(&responseBody)
	assert.True(t, strings.HasPrefix(responseBody.Result, "http://short.url/"))
	mockRepo.AssertExpectations(t)
}

type mockReader struct{}

func (m *mockReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("mock read error")
}

func TestJSONPostHandler_BadRequest(t *testing.T) {
	mockRepo := new(MockURLRepository)
	mockCookieManager := new(MockCookieManager)
	us := &URLShortener{
		URLRepository: mockRepo,
		CookieManager: mockCookieManager,
		BaseURL:       "http://short.url/",
	}

	req := httptest.NewRequest("POST", "/shorten", &mockReader{})
	w := httptest.NewRecorder()

	us.JSONPostHandler(w, req)

	// Assert
	res := w.Result()
	res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestJSONPostHandler_JSONUnmarshalError(t *testing.T) {
	mockRepo := new(MockURLRepository)
	mockCookieManager := new(MockCookieManager)
	us := &URLShortener{
		URLRepository: mockRepo,
		CookieManager: mockCookieManager,
		BaseURL:       "http://short.url/",
	}

	req := httptest.NewRequest("POST", "/shorten", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	us.JSONPostHandler(w, req)

	// Assert
	res := w.Result()
	res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestJSONPostHandler_InvalidURL(t *testing.T) {
	mockRepo := new(MockURLRepository)
	us := &URLShortener{
		URLRepository: mockRepo,
	}

	requestBody := RequestBody{URL: "invalid-url"}
	jsonBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/shorten", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	us.JSONPostHandler(w, req)

	// Assert
	res := w.Result()
	res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestJSONPostHandler_InternalServerErrorOnSave(t *testing.T) {
	mockRepo := new(MockURLRepository)
	mockCookieManager := new(MockCookieManager)
	us := &URLShortener{
		URLRepository: mockRepo,
		CookieManager: mockCookieManager,
		BaseURL:       "http://short.url/",
	}

	requestBody := RequestBody{URL: "http://example.com"}
	jsonBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/shorten", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Установка ожидания
	mockRepo.On("GetShortURL", mock.Anything).Return("", errors.New("short url not found"))
	mockRepo.On("Save", mock.Anything, requestBody.URL, mock.Anything).Return(errors.New("err"))
	mockCookieManager.On("GetActualCookieValue").Return("")

	// Act
	us.JSONPostHandler(w, req)

	// Assert
	res := w.Result()
	res.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestJSONBatchHandler_Success(t *testing.T) {
	mockRepo := new(MockURLRepository)
	mockCookieManager := new(MockCookieManager)
	us := &URLShortener{
		URLRepository: mockRepo,
		CookieManager: mockCookieManager,
		BaseURL:       "http://short.url/",
	}

	// Подготовка данных запроса
	requestBody := []BatchRequestBody{
		{CorrelationID: "1", OriginalURL: "http://example.com"},
		{CorrelationID: "2", OriginalURL: "http://anotherexample.com"},
	}
	jsonBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/batch/shorten", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Установка ожиданий на методы
	mockCookieManager.On("GetActualCookieValue").Return("")
	mockRepo.On("GetShortURL", "http://example.com").Return("", nil)
	mockRepo.On("GetShortURL", "http://anotherexample.com").Return("", nil)
	mockRepo.On("SaveBatch", mock.Anything).Return(nil)

	// Act
	us.JSONBatchHandler(w, req)

	res := w.Result()
	assert.Equal(t, http.StatusCreated, res.StatusCode)
	res.Body.Close()
	mockRepo.AssertExpectations(t)
}

func TestJSONBatchHandler_ReadError(t *testing.T) {
	us := &URLShortener{}

	req := httptest.NewRequest("POST", "/batch/shorten", nil) // nil вызываем ошибку
	w := httptest.NewRecorder()

	us.JSONBatchHandler(w, req)

	res := w.Result()
	res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode) // Ожидаем 400 Bad Request
}

func TestJSONBatchHandler_JSONUnmarshalError(t *testing.T) {
	us := &URLShortener{}

	req := httptest.NewRequest("POST", "/batch/shorten", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	us.JSONBatchHandler(w, req)

	res := w.Result()
	res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode) // Ожидаем 400 Bad Request
}

func TestJSONBatchHandler_ErrorOnGetShortURL(t *testing.T) {
	mockRepo := new(MockURLRepository)
	mockCookieManager := new(MockCookieManager)
	us := &URLShortener{
		URLRepository: mockRepo,
		CookieManager: mockCookieManager,
		BaseURL:       "http://short.url/",
	}

	// Подготовка данных запроса
	requestBody := []BatchRequestBody{
		{CorrelationID: "1", OriginalURL: "http://example.com"},
	}
	jsonBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/batch/shorten", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Установка ожидания на получение короткого URL, который вызывает ошибку
	mockCookieManager.On("GetActualCookieValue", mock.Anything).Return("")
	mockRepo.On("GetShortURL", "http://example.com").Return("", errors.New("get error"))
	mockRepo.On("SaveBatch", mock.Anything).Return(nil)

	// Act
	us.JSONBatchHandler(w, req)

	res := w.Result()
	assert.Equal(t, http.StatusCreated, res.StatusCode) // Ожидаем 201
	res.Body.Close()
}

func TestJSONBatchHandler_ErrorOnSaveBatch(t *testing.T) {
	mockRepo := new(MockURLRepository)
	mockCookieManager := new(MockCookieManager)
	us := &URLShortener{
		URLRepository: mockRepo,
		CookieManager: mockCookieManager,
		BaseURL:       "http://short.url/",
	}

	// Подготовка данных запроса
	requestBody := []BatchRequestBody{
		{CorrelationID: "1", OriginalURL: "http://example.com"},
	}
	jsonBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/batch/shorten", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Установка ожиданий
	mockCookieManager.On("GetActualCookieValue", mock.Anything).Return("")
	mockRepo.On("GetShortURL", "http://example.com").Return("", nil)                                         // Не найден
	mockRepo.On("Save", mock.Anything, "http://example.com", mock.Anything).Return(errors.New("save error")) // Ошибка при сохранении
	mockRepo.On("SaveBatch", mock.Anything).Return(errors.New("save error"))

	us.JSONBatchHandler(w, req)

	res := w.Result()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode) // Ожидаем 500
	res.Body.Close()
}

// Тест успешно получения URLs пользователя
func TestGetUserUrls_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCookieManager := new(MockCookieManager)
	us := &URLShortener{
		UserRepository: mockRepo,
		CookieManager:  mockCookieManager,
	}

	// Предполагаем, что GetActualCookieValue вернет "test_user_id"
	mockCookieManager.On("GetActualCookieValue").Return("test_user_id")

	// Подготовка данных
	expectedUrls := []storage.UserUrlsResponseBodyItem{
		{ShortURL: "shortURL1"},
		{ShortURL: "shortURL2"},
	}
	mockRepo.On("GetUserUrls", "test_user_id").Return(expectedUrls, nil) // Определяем, что должно быть возвращено

	// Создаем HTTP-запрос
	req := httptest.NewRequest("GET", "/user/urls", nil)
	w := httptest.NewRecorder()

	// Act
	us.GetUserUrls(w, req)

	// Assert
	res := w.Result()
	res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode) // Ожидаем статус 200 OK

	var responseBody []storage.UserUrlsResponseBodyItem
	_ = json.NewDecoder(res.Body).Decode(&responseBody)
	assert.Equal(t, expectedUrls, responseBody) // Проверка полученных данных

	mockRepo.AssertExpectations(t)
	mockCookieManager.AssertExpectations(t)
}

// Тест на ошибку во время получения URL
func TestGetUserUrls_Error(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCookieManager := new(MockCookieManager) // Создаем мок для CookieManager
	us := &URLShortener{
		UserRepository: mockRepo,
		CookieManager:  mockCookieManager,
	}

	mockCookieManager.On("GetActualCookieValue").Return("test_user_id")
	mockRepo.On("GetUserUrls", "test_user_id").Return([]storage.UserUrlsResponseBodyItem{}, errors.New("db error")) // Установка ожидания

	req := httptest.NewRequest("GET", "/user/urls", nil)
	w := httptest.NewRecorder()

	us.GetUserUrls(w, req)

	res := w.Result()
	res.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode) // Ожидаем статус 500 Internal Server Error

	mockRepo.AssertExpectations(t)
	mockCookieManager.AssertExpectations(t)
}

// Тест при отсутствии URL
func TestGetUserUrls_NoContent(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCookieManager := new(MockCookieManager) // Создаем мок для CookieManager
	us := &URLShortener{
		UserRepository: mockRepo,
		CookieManager:  mockCookieManager,
	}

	mockCookieManager.On("GetActualCookieValue").Return("test_user_id")
	mockRepo.On("GetUserUrls", "test_user_id").Return([]storage.UserUrlsResponseBodyItem{}, nil) // Пустой список

	req := httptest.NewRequest("GET", "/user/urls", nil)
	w := httptest.NewRecorder()

	us.GetUserUrls(w, req)

	res := w.Result()
	res.Body.Close()
	assert.Equal(t, http.StatusNoContent, res.StatusCode) // Ожидаем статус 204 No Content

	mockRepo.AssertExpectations(t)
	mockCookieManager.AssertExpectations(t)
}

func TestPostHandler_Success(t *testing.T) {
	mockRepo := new(MockURLRepository)
	mockCookieManager := new(MockCookieManager)
	us := &URLShortener{
		CookieManager: mockCookieManager,
		URLRepository: mockRepo,
		BaseURL:       "http://short.url/",
	}

	requestBody := `http://example.com`
	req := httptest.NewRequest("POST", "/shorten", bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "text/plain") // Если просто передаем URL в текстовом формате
	w := httptest.NewRecorder()

	// Устанавливаем ожидания
	mockCookieManager.On("GetActualCookieValue").Return("")
	mockRepo.On("GetShortURL", requestBody).Return("", errors.New("short url not found")) // URL не найден
	mockRepo.On("Save", mock.Anything, requestBody, mock.Anything).Return(nil)            // Успешно сохранить

	// Act
	us.PostHandler(w, req)

	// Assert
	res := w.Result()
	res.Body.Close()
	assert.Equal(t, http.StatusCreated, res.StatusCode)

	var responseBody JSONResponseBody
	_ = json.NewDecoder(res.Body).Decode(&responseBody)
	assert.True(t, strings.HasPrefix(responseBody.Result, "")) // Проверка результата

	mockRepo.AssertExpectations(t)
}

func TestPostHandler_ReadError(t *testing.T) {
	us := &URLShortener{}

	req := httptest.NewRequest("POST", "/shorten", nil) // nil как тело запроса
	w := httptest.NewRecorder()

	// Act
	us.PostHandler(w, req)

	// Assert
	res := w.Result()
	res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode) // Ожидаем 400 Bad Request
}

func TestPostHandler_InvalidURL(t *testing.T) {
	us := &URLShortener{}

	// Подготовка некорректного тела запроса
	requestBody := `invalid-url`
	req := httptest.NewRequest("POST", "/shorten", bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	// Act
	us.PostHandler(w, req)

	// Assert
	res := w.Result()
	res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode) // Ожидаем 400 Bad Request
}

// TestDeleteUserUrlsBatch_Success - тестирует успешное удаление URLs
func TestDeleteUserUrlsBatch_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCookieManager := new(MockCookieManager)

	us := &URLShortener{
		UserRepository: mockRepo,
		CookieManager:  mockCookieManager,
	}

	// Устанавливаем ожидания для CookieManager
	mockCookieManager.On("GetActualCookieValue").Return("test_user_id")

	// Подготовка тестовых данных
	shortURLs := []string{"shortURL1", "shortURL2", "shortURL3", "shortURL4", "shortURL5"}

	// Устанавливаем ожидания на удаление
	mockRepo.On("DeleteUserUrls", "test_user_id", mock.Anything).Return(nil).Once()

	us.DeleteUserUrlsBatch(shortURLs)

	mockRepo.AssertExpectations(t)
}

// TestDeleteUserUrlsBatch_DeleteError - тестирует поведение при ошибке удаления
func TestDeleteUserUrlsBatch_DeleteError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCookieManager := new(MockCookieManager)

	us := &URLShortener{
		UserRepository: mockRepo,
		CookieManager:  mockCookieManager,
	}

	// Устанавливаем ожидания для CookieManager
	mockCookieManager.On("GetActualCookieValue").Return("test_user_id")

	// Подготовка тестовых данных
	shortURLs := []string{"shortURL1", "shortURL2", "shortURL3"}

	// Устанавливаем ожидание на удаление с ошибкой
	mockRepo.On("DeleteUserUrls", mock.Anything, mock.Anything).Return(errors.New("delete error")).Once()

	us.DeleteUserUrlsBatch(shortURLs)

	mockRepo.AssertExpectations(t)
}

// TestDeleteUserUrlsBatch_WithBatches - тестирует поведение с несколькими батчами
func TestDeleteUserUrlsBatch_WithBatches(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCookieManager := new(MockCookieManager)

	us := &URLShortener{
		UserRepository: mockRepo,
		CookieManager:  mockCookieManager,
	}

	// Устанавливаем ожидания для CookieManager
	mockCookieManager.On("GetActualCookieValue").Return("test_user_id")

	// Подготовка тестовых данных, больше чем 100 элементов
	shortURLs := make([]string, 150)
	for i := 0; i < 150; i++ {
		shortURLs[i] = "shortURL" + strconv.Itoa(i) // Правильная конвертация int в string
	}

	// Устанавливаем ожидание на удаление для первых 100
	mockRepo.On("DeleteUserUrls", "test_user_id", shortURLs[:100]).Return(nil).Once()
	// Устанавливаем ожидание на удаление для оставшихся 50
	mockRepo.On("DeleteUserUrls", "test_user_id", shortURLs[100:150]).Return(nil).Once()

	// Act
	us.DeleteUserUrlsBatch(shortURLs)

	// Assert
	mockRepo.AssertExpectations(t)
}

func TestDeleteUserUrls_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCookieManager := new(MockCookieManager)

	us := &URLShortener{
		UserRepository: mockRepo,
		CookieManager:  mockCookieManager,
		RemoveChan:     make(chan string, 10), // Буферизированный канал
	}

	shortURLs := []string{"shortURL1", "shortURL2", "shortURL3"}
	body, _ := json.Marshal(shortURLs)

	// Создаем тестовый HTTP-запрос
	req := httptest.NewRequest("DELETE", "/user/urls", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	us.DeleteUserUrls(w, req)

	// Assert
	res := w.Result()
	res.Body.Close()
	assert.Equal(t, http.StatusAccepted, res.StatusCode) // Ожидаем статус 202 Accepted

	// Проверяем, что короткие URL отправлены в канал RemoveChan
	close(us.RemoveChan)
	for shortURL := range us.RemoveChan {
		assert.Contains(t, shortURLs, shortURL) // Проверяем, что URL находится в нашем списке
	}

	mockRepo.AssertExpectations(t)
	mockCookieManager.AssertExpectations(t)
}

func TestDeleteUserUrls_JSONDecodeError(t *testing.T) {
	us := &URLShortener{}

	// Создаем тестовый HTTP-запрос с некорректным JSON
	req := httptest.NewRequest("DELETE", "/user/urls", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	us.DeleteUserUrls(w, req)

	// Assert
	res := w.Result()
	res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode) // Ожидаем статус 400 Bad Request
}

func TestDeleteUserUrls_EmptyBody(t *testing.T) {
	us := &URLShortener{}

	// Создаем тестовый HTTP-запрос без тела
	req := httptest.NewRequest("DELETE", "/user/urls", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	us.DeleteUserUrls(w, req)

	// Assert
	res := w.Result()
	res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode) // Ожидаем статус 400 Bad Request
}
