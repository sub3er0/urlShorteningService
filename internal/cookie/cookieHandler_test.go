package cookie

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sub3er0/urlShorteningService/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserStorage - мок для UserStorageInterface
type MockUserStorage struct {
	mock.Mock
}

// IsUserExist - реализует метод интерфейса UserStorageInterface
func (m *MockUserStorage) IsUserExist(uniqueID string) bool {
	args := m.Called(uniqueID)
	return args.Bool(0)
}

// SaveUser - реализует метод интерфейса UserStorageInterface
func (m *MockUserStorage) SaveUser(uniqueID string) error {
	args := m.Called(uniqueID)
	return args.Error(0)
}

// GetUserUrls - реализует метод интерфейса UserStorageInterface
func (m *MockUserStorage) GetUserUrls(uniqueID string) ([]storage.UserUrlsResponseBodyItem, error) {
	args := m.Called(uniqueID)
	return args.Get(0).([]storage.UserUrlsResponseBodyItem), args.Error(1)
}

// DeleteUserUrls - реализует метод интерфейса UserStorageInterface
func (m *MockUserStorage) DeleteUserUrls(uniqueID string, shortURLs []string) error {
	args := m.Called(uniqueID, shortURLs)
	return args.Error(0)
}

// Init - реализует метод интерфейса UserStorageInterface
func (m *MockUserStorage) Init(connectionString string) error {
	args := m.Called(connectionString)
	return args.Error(0)
}

// Close - реализует метод интерфейса UserStorageInterface
func (m *MockUserStorage) Close() {
	m.Called()
}

func TestCookieHandler_NewUserCreated(t *testing.T) {
	// Arrange
	mockStorage := new(MockUserStorage)
	cm := &CookieManager{
		Storage: mockStorage,
	}

	request := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	// Установка ожиданий для методов хранилища
	mockStorage.On("SaveUser", mock.Anything).Return(nil)

	// Act
	handler := cm.CookieHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	handler.ServeHTTP(recorder, request)

	// Assert
	res := recorder.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Проверка, что пользователь сохранен и новая кука установлена
	mockStorage.AssertExpectations(t)
}

func TestCookieHandler_ExistingUser(t *testing.T) {
	// Arrange
	mockStorage := new(MockUserStorage)
	cm := &CookieManager{
		Storage: mockStorage,
	}

	cookieValue := "someUserID." + signCookie("someUserID") // Создаем куки
	request := httptest.NewRequest("GET", "/", nil)
	request.AddCookie(&http.Cookie{Name: "user_info", Value: cookieValue})
	recorder := httptest.NewRecorder()

	// Установка ожиданий для методов хранилища
	mockStorage.On("IsUserExist", "someUserID").Return(true)

	// Act
	handler := cm.CookieHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	handler.ServeHTTP(recorder, request)

	// Assert
	res := recorder.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Проверка, что метод IsUserExist был вызван
	mockStorage.AssertExpectations(t)
}

func TestCookieHandler_InvalidCookie(t *testing.T) {
	// Arrange
	mockStorage := new(MockUserStorage)
	cm := &CookieManager{
		Storage: mockStorage,
	}

	mockStorage.On("SaveUser", mock.Anything).Return(nil)

	request := httptest.NewRequest("GET", "/", nil) // Запрос без куки
	recorder := httptest.NewRecorder()

	// Act
	handler := cm.CookieHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	handler.ServeHTTP(recorder, request)

	// Assert
	res := recorder.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Проверка, что метод SaveUser был вызван для нового пользователя
	mockStorage.AssertExpectations(t)
}
