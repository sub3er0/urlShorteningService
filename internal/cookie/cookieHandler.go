package cookie

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/sub3er0/urlShorteningService/internal/storage"
)

// CookieManager управляет аутентификацией и обработкой куки в приложении.
// Он предоставляет методы для установки, проверки и получения значений куки.
type CookieManager struct {
	// Storage используется для взаимодействия с хранилищем пользователей.
	Storage storage.UserStorageInterface

	// ActualCookieValue содержит текущее значение куки аутентификации пользователя.
	ActualCookieValue string
}

// CookieManagerInterface определяет методы для работы с куками в приложении.
// Этот интерфейс позволяет управлять куками и реализовывать middleware для аутентификации.
type CookieManagerInterface interface {
	// CookieHandler оборачивает HTTP-обработчик для управления кукми.
	// Возвращает обработчик, который изменён для работы с куками.
	CookieHandler(h http.Handler) http.Handler

	// AuthMiddleware оборачивает HTTP-обработчик для проверки аутентификации пользователя.
	// Внутри проверяет наличие и корректность куки, а также существование пользователя.
	// Если аутентификация не пройдена, возвращает статус 401 Unauthorized.
	AuthMiddleware(h http.Handler) http.Handler

	// GetActualCookieValue возвращает значение актуальной куки для текущего пользователя.
	GetActualCookieValue() string
}

// GetActualCookieValue возвращает значение актуальной куки для текущего пользователя.
// Возвращает строку, представляющую актуальное значение куки.
func (cm *CookieManager) GetActualCookieValue() string {
	return cm.ActualCookieValue
}

var (
	secretKey  = []byte("secret_key")
	cookieName = "user_info"
)

// signCookie вычисляет HMAC-подпись для данных.
// Используется для проверки подлинности куки.
func signCookie(data string) string {
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(data))
	signature := h.Sum(nil)

	return base64.RawURLEncoding.EncodeToString(signature)
}

// verifyCookie проверяет корректность куки, сравнивая подпись с ожидаемой.
func verifyCookie(str string) bool {
	parts := strings.Split(str, ".")
	if len(parts) != 2 {
		return false
	}

	dataStr := parts[0]
	sigStr := parts[1]
	expected := signCookie(dataStr)

	return sigStr == expected
}

// getUserIDFromCookie извлекает идентификатор пользователя из куки.
func getUserIDFromCookie(str string) (string, bool) {
	parts := strings.Split(str, ".")
	if len(parts) != 2 {
		return "", false
	}

	return parts[0], true
}

// generateUserID генерирует уникальный идентификатор пользователя.
func generateUserID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// CookieHandler оборачивает HTTP-обработчик, добавляя логику работы с куками.
func (cm *CookieManager) CookieHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieName)
		createNewCookie := false
		var userID string

		if err != nil || !verifyCookie(cookie.Value) {
			createNewCookie = true
		} else {
			userID, _ = getUserIDFromCookie(cookie.Value)
			isUserExist := cm.Storage.IsUserExist(userID)

			if isUserExist {
				createNewCookie = false
			} else {
				createNewCookie = true
			}
		}

		if createNewCookie {
			userID = generateUserID()
			newCookieValue := userID + "." + signCookie(userID)
			cm.Storage.SaveUser(userID)
			http.SetCookie(w, &http.Cookie{
				Name:     cookieName,
				Value:    newCookieValue,
				Path:     "/",
				Expires:  time.Now().AddDate(10, 0, 0),
				HttpOnly: true,
				Secure:   false,
			})
		}

		cm.ActualCookieValue = userID
		h.ServeHTTP(w, r)
	})
}
