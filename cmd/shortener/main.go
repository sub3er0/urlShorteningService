package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type URLShortener struct {
	urls map[string]string
}

func (us *URLShortener) GetHandler(w http.ResponseWriter, r *http.Request) {
	// этот обработчик принимает только запросы, отправленные методом GET
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
	// продолжаем обработку запроса
	// ...
}

func (us *URLShortener) PostHandler(w http.ResponseWriter, r *http.Request) {
	// этот обработчик принимает только запросы, отправленные методом GET
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	shortKey := generateShortKey()
	us.urls[shortKey] = r.RequestURI
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	shortenedURL := fmt.Sprintf("http://localhost:8080/%s", shortKey)
	fmt.Fprintf(w, shortenedURL)
}

func generateShortKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLength = 6

	rand.Seed(time.Now().UnixNano())
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

	http.HandleFunc("/{id}}", shortener.GetHandler)
	http.HandleFunc("/", shortener.PostHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
