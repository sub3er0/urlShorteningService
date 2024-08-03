package shortener

import (
	"encoding/json"
	"github.com/sub3er0/urlShorteningService/internal/storage"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
)

// URLShortener Структура URLShortener, использующая интерфейс хранения
type URLShortener struct {
	Storage       storage.URLStorage
	ServerAddress string
	BaseURL       string
	DataStorage   storage.DataStorageInterface
}

type JSONResponseBody struct {
	Result string `json:"result"`
}

type RequestBody struct {
	URL string `json:"url"`
}

// GetURL Реализация функции получения URL
func (us *URLShortener) GetURL(shortURL string) (string, bool) {
	return us.Storage.GetURL(shortURL)
}

// SetURL Реализация функции сохранения URL
func (us *URLShortener) SetURL(shortURL, longURL string) error {
	return us.Storage.Set(shortURL, longURL)
}

func (us *URLShortener) getShortURL(URL string) (string, bool) {
	return us.Storage.GetShortURL(URL)
}

func (us *URLShortener) GetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}

	id := r.PathValue("id")

	storedURL, ok := us.Storage.GetURL(id)

	if ok {
		w.Header().Set("Location", storedURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
}

func (us *URLShortener) JSONPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var requestBody RequestBody
	err = json.Unmarshal(body, &requestBody)

	if err != nil {
		return
	}

	bodyURL, err := url.ParseRequestURI(requestBody.URL)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shortKey, err := us.getShortKey(bodyURL.String())

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var responseBody JSONResponseBody
	responseBody.Result = shortKey

	err = us.buildJSONResponse(w, responseBody)

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
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
	shortKey, err := us.getShortKey(postURL)

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusBadRequest)
		return
	}

	us.buildResponse(w, shortKey)

	defer func(Body io.ReadCloser) {
		err := Body.Close()

		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}(r.Body)

	return
}

func (us *URLShortener) getShortKey(postURL string) (string, error) {
	shortKey, ok := us.getShortURL(postURL)

	if ok {
		return shortKey, nil
	}

	shortKey = generateShortKey()
	err := us.SetURL(shortKey, postURL)

	if err != nil {
		return "", err
	}

	FileStorageRowStruct := storage.DataStorageRow{
		ID:       us.Storage.GetURLCount(),
		ShortURL: shortKey,
		URL:      postURL,
	}
	err = us.DataStorage.Save(FileStorageRowStruct)

	if err != nil {
		return "", err
	}

	return shortKey, nil
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

func (us *URLShortener) buildResponse(w http.ResponseWriter, shortKey string) {
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	if len(us.BaseURL) > 0 && us.BaseURL[len(us.BaseURL)-1] != '/' {
		us.BaseURL = us.BaseURL + "/"
	}

	_, err := w.Write([]byte(us.BaseURL + shortKey))

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (us *URLShortener) buildJSONResponse(w http.ResponseWriter, response JSONResponseBody) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if len(us.BaseURL) > 0 && us.BaseURL[len(us.BaseURL)-1] != '/' {
		us.BaseURL = us.BaseURL + "/"
	}

	response.Result = us.BaseURL + response.Result
	jsonData, err := json.Marshal(response)

	if err != nil {
		log.Printf("Serialization fail: %v", err)
		return err
	}

	_, err = w.Write(jsonData)

	if err != nil {
		log.Printf("Write data error: %v", err)
		return err
	}

	return nil
}

func (us *URLShortener) LoadData() error {
	DataStorageRows, err := us.DataStorage.LoadData()

	if err != nil {
		return err
	}

	for _, dataStorageRow := range DataStorageRows {
		err := us.Storage.Set(dataStorageRow.ShortURL, dataStorageRow.URL)

		if err != nil {
			return err
		}
	}

	return nil
}
