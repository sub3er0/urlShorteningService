package cookie

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"github.com/sub3er0/urlShorteningService/internal/storage"
	"net/http"
	"strings"
	"time"
)

type CookieManager struct {
	Storage           storage.URLStorage
	ActualCookieValue string
}

var (
	secretKey  = []byte("secret_key")
	cookieName = "user_info"
)

func signCookie(data string) string {
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(data))
	signature := h.Sum(nil)

	return base64.RawURLEncoding.EncodeToString(signature)
}

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

func getUserIDFromCookie(str string) (string, bool) {
	parts := strings.Split(str, ".")
	if len(parts) != 2 {
		return "", false
	}

	return parts[0], true
}

func splitCookieData(data string) []string {
	//decodedData, err := base64.RawURLEncoding.DecodeString(data)
	//
	//if err != nil {
	//	return nil
	//}

	return strings.Split(string(data), ".")
}

func generateUserID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (cm *CookieManager) CookieHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieName)
		createNewCookie := false
		var userID string

		if err != nil {
			createNewCookie = true
		} else if !verifyCookie(cookie.Value) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		} else {
			userID, _ = getUserIDFromCookie(cookie.Value)

			if userID == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

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
