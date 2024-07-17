package shortener

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
)

type URLShortener struct {
	Urls          map[string]string
	ServerAddress string
	BaseURL       string
}

func (us *URLShortener) GetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}

	id := r.PathValue("id")

	for k, v := range us.Urls {
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
	shortKey := us.getShortKey(postURL)
	us.buildResponse(w, r, shortKey)

	defer func(Body io.ReadCloser) {
		err := Body.Close()

		if err != nil {
			log.Fatalf("Error while initializing config: %v", err)
		}
	}(r.Body)
}

func (us *URLShortener) getShortKey(postURL string) string {
	for k, v := range us.Urls {
		if v == postURL {
			return k
		}
	}

	shortKey := generateShortKey()
	us.Urls[shortKey] = postURL

	return shortKey
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

func (us *URLShortener) buildResponse(w http.ResponseWriter, r *http.Request, shortKey string) {
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	if len(us.BaseURL) > 0 && us.BaseURL[len(us.BaseURL)-1] != '/' {
		us.BaseURL = us.BaseURL + "/"
	}

	w.Write([]byte(us.BaseURL + shortKey))
}
