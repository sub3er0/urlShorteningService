package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/sub3er0/urlShorteningService/cmd/config"
	"io"
	"math/rand"
	"net/http"
	"net/url"
)

type URLShortener struct {
	urls          map[string]string
	ServerAddress string
	BaseURL       string
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
			us.buildResponse(w, r, k)
			return
		}
	}

	shortKey := generateShortKey()
	us.urls[shortKey] = postURL
	us.buildResponse(w, r, shortKey)
}

func (us *URLShortener) buildResponse(w http.ResponseWriter, r *http.Request, shortKey string) {
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(us.BaseURL + shortKey))
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
	cfg, err := config.InitConfig()

	if err != nil {
		return
	}

	shortener := &URLShortener{
		urls:          make(map[string]string),
		ServerAddress: cfg.ServerAddress,
		BaseURL:       cfg.BaseURL,
	}

	r := chi.NewRouter()
	r.Post("/", shortener.PostHandler)
	r.Get("/{id}", shortener.GetHandler)
	err = http.ListenAndServe(cfg.ServerAddress, r)

	if err != nil {
		return
	}
}
