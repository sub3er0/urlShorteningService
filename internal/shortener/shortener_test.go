package shortener

import (
	"bytes"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/sub3er0/urlShorteningService/internal/cookie"
	"github.com/sub3er0/urlShorteningService/internal/storage"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJsonPostHandler_InvalidURL(t *testing.T) {
	us := &URLShortener{
		Storage: &storage.InMemoryStorage{
			Urls: map[string]string{},
		},
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080/",
		CookieManager: &cookie.CookieManager{ActualCookieValue: "test"},
	}
	var requestBody RequestBody
	requestBody.URL = "invalid-url"
	jsonBody, err := json.Marshal(requestBody)

	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(jsonBody))

	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	us.JSONPostHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code, "Invalid URL")
}

func TestURLShortener_JsonPostHandlerExistedUrl(t *testing.T) {
	us := &URLShortener{
		Storage: &storage.InMemoryStorage{
			Urls: map[string]string{"shortURL": "https://www.example.com"},
		},
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080/",
		CookieManager: &cookie.CookieManager{ActualCookieValue: "test"},
	}

	var requestBody RequestBody
	requestBody.URL = "https://www.example.com"
	jsonBody, err := json.Marshal(requestBody)

	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost/api/shorten", bytes.NewReader(jsonBody))
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	us.JSONPostHandler(w, req)
	assert.Equal(t, http.StatusConflict, w.Code, "Incorrect status code")

	body := w.Body.String()
	assert.Contains(t, body, "\"result\":\"http://localhost:8080/shortURL\"", "Body must looks like \"result\":\"http://localhost\"")
}

func TestURLShortener_PostHandler(t *testing.T) {
	us := &URLShortener{
		Storage:       &storage.InMemoryStorage{Urls: make(map[string]string)},
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080/",
		CookieManager: &cookie.CookieManager{ActualCookieValue: "test"},
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost/", bytes.NewReader([]byte("https://www.example.com")))
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	us.PostHandler(w, req)
	assert.Equal(t, http.StatusCreated, w.Code, "Invalid status code")

	body := w.Body.String()
	assert.Contains(t, body, "http://localhost", "Invalid url in response body")
}

func TestGetHandler_ValidRequest(t *testing.T) {
	us := &URLShortener{
		Storage: &storage.InMemoryStorage{
			Urls: map[string]string{"shortURL": "https://www.example.com"},
		},
	}
	_, err := http.NewRequest(http.MethodGet, "/shortURL", nil)

	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	originalURL, ok := us.Storage.GetURL("shortURL")

	if ok {
		w.Header().Set("Location", originalURL.URL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code, "Invalid status code")

	location := w.Header().Get("Location")
	assert.Equal(t, "https://www.example.com", location, "Invalid Location header value")
}

func TestPostHandler_InvalidMethod(t *testing.T) {
	router := chi.NewRouter()

	us := &URLShortener{
		Storage: &storage.InMemoryStorage{
			Urls: map[string]string{"shortURL": "https://www.example.com"},
		},
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080/",
		CookieManager: &cookie.CookieManager{ActualCookieValue: "test"},
	}

	router.Post("/", us.PostHandler)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code, "Invalid status code")
}

func TestPostHandler_InvalidURL(t *testing.T) {
	us := &URLShortener{
		Storage: &storage.InMemoryStorage{
			Urls: map[string]string{},
		},
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080/",
		CookieManager: &cookie.CookieManager{ActualCookieValue: "test"},
	}
	body := []byte("invalid-url")
	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewReader(body))

	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	us.PostHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code, "Invalid URL")
}

func TestURLShortener_PostHandlerExistedUrl(t *testing.T) {
	us := &URLShortener{
		Storage: &storage.InMemoryStorage{
			Urls: map[string]string{"shortURL": "https://www.example.com"},
		},
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080/",
		CookieManager: &cookie.CookieManager{ActualCookieValue: "test"},
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost/", bytes.NewReader([]byte("https://www.example.com")))
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	us.PostHandler(w, req)
	assert.Equal(t, http.StatusConflict, w.Code, "Incorrect status code")

	body := w.Body.String()
	assert.Contains(t, body, "http://localhost:8080/shortURL", "Body must looks like http://localhost:8080/shortURL")
}

func TestURLShortener_GetShortKeyExist(t *testing.T) {
	us := &URLShortener{
		Storage: &storage.InMemoryStorage{
			Urls: map[string]string{"shortURL": "https://www.example.com"},
		},
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080/",
		CookieManager: &cookie.CookieManager{ActualCookieValue: "test"},
	}
	shortKey, err := us.getShortURL("https://www.example.com")

	if err != nil {
		assert.Equal(t, "shortURL", shortKey, "Incorrect Short URL")
	}
}

func TestURLShortener_PostHandlerTable(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(us *URLShortener)
		us      *URLShortener
		wantErr bool
	}{
		{
			name: "json post handler",
			us: &URLShortener{
				Storage:       &storage.InMemoryStorage{Urls: make(map[string]string)},
				ServerAddress: "localhost:8080",
				BaseURL:       "http://localhost:8080/",
				CookieManager: &cookie.CookieManager{ActualCookieValue: "test"},
			},
			prepare: func(us *URLShortener) {
				var requestBody RequestBody
				requestBody.URL = "https://www.example.com"
				jsonBody, err := json.Marshal(requestBody)

				if err != nil {
					return
				}

				req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/shorten", bytes.NewReader(jsonBody))
				assert.NoError(t, err)

				w := httptest.NewRecorder()
				us.JSONPostHandler(w, req)
				assert.Equal(t, http.StatusCreated, w.Code, "Invalid status code")

				body := w.Body.String()
				assert.Contains(t, body, "http://localhost:8080", "Invalid url in response body")
			},
		},
		{
			name: "invalid method",
			us: &URLShortener{
				Storage: &storage.InMemoryStorage{
					Urls: map[string]string{"shortURL": "https://www.example.com"},
				},
				ServerAddress: "localhost:8080",
				BaseURL:       "http://localhost:8080/",
				CookieManager: &cookie.CookieManager{ActualCookieValue: "test"},
			},
			prepare: func(us *URLShortener) {
				router := chi.NewRouter()
				router.Post("/api/shorten", us.JSONPostHandler)

				req, err := http.NewRequest(http.MethodGet, "/api/shorten", nil)
				if err != nil {
					t.Fatal(err)
				}

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				assert.Equal(t, http.StatusMethodNotAllowed, w.Code, "Invalid status code")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			if tt.prepare != nil {
				tt.prepare(tt.us)
			}
		})
	}
}
