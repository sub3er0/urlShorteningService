package main

import (
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

type URLShortener struct {
	urls map[string]string
}

func (us *URLShortener) GetHandler(w http.ResponseWriter, r *http.Request) {
	// этот обработчик принимает только запросы, отправленные методом GET
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}

	id := r.PathValue("id")

	for k, v := range us.urls {
		if k == id {
			resp, err := json.Marshal(v)

			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			w.WriteHeader(http.StatusTemporaryRedirect)
			w.Write(resp)
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

	postUrl := u.String()

	for k, v := range us.urls {
		if v == postUrl {
			buildResponse(w, r, k)
			return
		}
	}

	shortKey := generateShortKey()
	us.urls[shortKey] = postUrl
	buildResponse(w, r, shortKey)
}

func buildResponse(w http.ResponseWriter, r *http.Request, shortKey string) {
	resp, err := json.Marshal(r.Host + "/" + shortKey)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
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

	http.HandleFunc("/{id}", shortener.GetHandler)
	http.HandleFunc("/", shortener.PostHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
