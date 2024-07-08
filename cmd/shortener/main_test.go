package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestURLShortener_PostHandler(t *testing.T) {
	us := &URLShortener{
		urls:          map[string]string{},
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080/",
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost/", bytes.NewReader([]byte("https://www.example.com")))
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	us.PostHandler(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	body := w.Body.String()
	assert.Contains(t, body, "http://localhost")
}

func TestGetHandler_ValidRequest(t *testing.T) {
	us := &URLShortener{
		urls: map[string]string{
			"shortURL": "https://www.example.com",
		},
	}
	_, err := http.NewRequest(http.MethodGet, "/shortURL", nil)

	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	originalURL := us.urls["shortURL"]
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)

	location := w.Header().Get("Location")
	assert.Equal(t, "https://www.example.com", location)
}

func TestPostHandler_InvalidMethod(t *testing.T) {
	us := &URLShortener{
		urls:          map[string]string{},
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080/",
	}
	req, err := http.NewRequest(http.MethodGet, "/", nil)

	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	us.PostHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	body := w.Body.String()
	assert.Equal(t, "Only POST requests are allowed!\n", body)
}

func TestPostHandler_InvalidURL(t *testing.T) {
	us := &URLShortener{
		urls:          map[string]string{},
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080/",
	}
	body := []byte("invalid-url")
	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewReader(body))

	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	us.PostHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
