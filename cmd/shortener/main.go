package main

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
)

type URLShortener struct {
	urls map[string]string
}

func (us *URLShortener) GetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}

	id := r.PathValue("id")

	for k, v := range us.urls {
		if k == id {
			w.Header().Set("Location", v)
			w.WriteHeader(http.StatusTemporaryRedirect)
			return
		}
	}
}

func (us *URLShortener) PostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	u, err := url.ParseRequestURI(string(body))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	postURL := u.String()

	for k, v := range us.urls {
		if v == postURL {
			buildResponse(w, r, k)
			return
		}
	}

	shortKey := generateShortKey()
	us.urls[shortKey] = postURL
	buildResponse(w, r, shortKey)
}

func buildResponse(w http.ResponseWriter, r *http.Request, shortKey string) {
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("http://" + r.Host + "/" + shortKey))
}

func generateShortKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLength = 6

	shortKey := make([]byte, keyLength)

	for i := range shortKey {
		shortKey[i] = charset[rand.Intn(len(charset))]
	}

	return string(shortKey)
}

func main() {
	shortener := &URLShortener{
		urls: make(map[string]string),
	}

	http.HandleFunc("/{id}", shortener.GetHandler)
	http.HandleFunc("/", shortener.PostHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
