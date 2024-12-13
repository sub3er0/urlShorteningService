package cookie

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_MissingCookie(t *testing.T) {
	mockStorage := new(MockUserStorage)
	cm := &CookieManager{Storage: mockStorage}

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Создаем мидлвар
	handler := cm.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode) // Ожидаем статус 401 Unauthorized
}

func TestAuthMiddleware_InvalidCookie(t *testing.T) {
	mockStorage := new(MockUserStorage)
	cm := &CookieManager{
		Storage: mockStorage,
	}

	// Создаем запрос с недопустимой кукой
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: "invalid.cookie.value"})
	w := httptest.NewRecorder()

	// Act
	handler := cm.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(w, req)

	// Assert
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode) // Ожидаем статус 401 Unauthorized
}

func TestAuthMiddleware_UserNotExists(t *testing.T) {
	mockStorage := new(MockUserStorage)
	cm := &CookieManager{
		Storage: mockStorage,
	}

	// Создаем куку с правильным значением
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: "userID." + SignCookie("userID")})
	w := httptest.NewRecorder()

	// Устанавливаем ожидание для метода IsUserExist
	mockStorage.On("IsUserExist", "userID").Return(false) // Пользователь не существует

	// Act
	handler := cm.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(w, req)

	// Assert
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode) // Ожидаем статус 401 Unauthorized

	// Проверка ожиданий
	mockStorage.AssertExpectations(t)
}

func TestAuthMiddleware_Success(t *testing.T) {
	mockStorage := new(MockUserStorage)
	cm := &CookieManager{
		Storage: mockStorage,
	}

	// Создаем куку с правильным значением
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: "userID." + SignCookie("userID")})
	w := httptest.NewRecorder()

	// Устанавливаем ожидание для метода IsUserExist
	mockStorage.On("IsUserExist", "userID").Return(true) // Пользователь существует

	// Act
	handler := cm.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(w, req)

	// Assert
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode) // Ожидаем статус 200 OK

	// Проверка ожиданий
	mockStorage.AssertExpectations(t)
}
