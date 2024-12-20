package cookie

import (
	"net/http"
)

// AuthMiddleware оборачивает HTTP-обработчик для проверки аутентификации пользователя.
// Этот мидлвар проверяет наличие куки с именем user_info и ее валидность.
func (cm *CookieManager) AuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieName)
		if err != nil || !verifyCookie(cookie.Value) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID, _ := getUserIDFromCookie(cookie.Value)
		isUserExist := cm.Storage.IsUserExist(userID)

		if !isUserExist {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		h.ServeHTTP(w, r)
	})
}
