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

	req, err := http.NewRequest(http.MethodGet, "/shortURL", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	us.GetHandler(rr, req)

	if status := rr.Code; status != http.StatusTemporaryRedirect {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusTemporaryRedirect)
	}

	if location := rr.Header().Get("Location"); location != "https://www.example.com" {
		t.Errorf("handler returned wrong location header: got %v want %v",
			location, "https://www.example.com")
	}
}
