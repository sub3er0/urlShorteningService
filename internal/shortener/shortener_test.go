package shortener

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/sub3er0/urlShorteningService/internal/storage"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestURLShortener_JsonPostHandler(t *testing.T) {
	us := &URLShortener{
		Storage:       &storage.InMemoryStorage{Urls: make(map[string]string)},
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080/",
	}

	var requestBody RequestBody
	requestBody.URL = "https://www.example.com"
	jsonBody, err := json.Marshal(requestBody)

	if err != nil {
		fmt.Println("Deserialization fail:", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/shorten", bytes.NewReader(jsonBody))
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	us.JSONPostHandler(w, req)
	assert.Equal(t, http.StatusCreated, w.Code, "Invalid status code")

	body := w.Body.String()
	assert.Contains(t, body, "http://localhost:8080", "Invalid url in response body")
}

func TestJsonPostHandler_InvalidMethod(t *testing.T) {
	us := &URLShortener{
		Storage: &storage.InMemoryStorage{
			Urls: map[string]string{"shortURL": "https://www.example.com"},
		},
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080/",
	}
	req, err := http.NewRequest(http.MethodGet, "/api/shorten", nil)

	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	us.JSONPostHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code, "Invalid status code")

	body := w.Body.String()
	assert.Equal(t, "Only POST requests are allowed!\n", body, "Invalid response message")
}

func TestJsonPostHandler_InvalidURL(t *testing.T) {
	us := &URLShortener{
		Storage: &storage.InMemoryStorage{
			Urls: map[string]string{},
		},
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080/",
	}
	var requestBody RequestBody
	requestBody.URL = "invalid-url"
	jsonBody, err := json.Marshal(requestBody)

	if err != nil {
		fmt.Println("Deserialization fail:", err)
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
	}

	var requestBody RequestBody
	requestBody.URL = "https://www.example.com"
	jsonBody, err := json.Marshal(requestBody)

	if err != nil {
		fmt.Println("Deserialization fail:", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost/api/shorten", bytes.NewReader(jsonBody))
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	us.JSONPostHandler(w, req)
	assert.Equal(t, http.StatusCreated, w.Code, "Incorrect status code")

	body := w.Body.String()
	assert.Contains(t, body, "\"result\":\"http://localhost:8080/shortURL\"", "Body must looks like \"result\":\"http://localhost\"")
}
