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
		urls: map[string]string{},
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost/", bytes.NewReader([]byte("https://www.example.com")))
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	us.PostHandler(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	body := w.Body.String()
	assert.Contains(t, body, "http://localhost")
}

func TestGetHandler(t *testing.T) {
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
