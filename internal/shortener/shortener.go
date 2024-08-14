package shortener

import (
	"encoding/json"
	"fmt"
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
}

type JSONResponseBody struct {
	Result string `json:"result"`
}

type RequestBody struct {
	URL string `json:"url"`
}

type BatchRequestBody struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchResponseBodyItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
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
		parsedURL, err := url.Parse(storedURL)

		if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
			w.Header().Set("Location", storedURL)
			w.WriteHeader(http.StatusOK)
			return
		}

		w.Header().Set("Location", storedURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	} else {
		http.Error(w, "NotFound", http.StatusNotFound)
	}
}

func expand(originalURL string) (string, error) {
	// Проверяем валидность URL
	parsedURL, err := url.Parse(originalURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return "", fmt.Errorf("невалидный URL: %s", originalURL)
	}

	// Создаем HTTP-клиент с отключением редиректов
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return fmt.Errorf("редиректы заблокированы")
		},
	}

	response, err := client.Get(originalURL)
	if err != nil {
		return "", fmt.Errorf("ошибка при выполнении GET-запроса: %w", err)
	}
	defer response.Body.Close()

	// Дополнительная обработка ответа...

	return "result", nil
}

func (us *URLShortener) PingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}

	ok := us.Storage.Ping()

	if ok {
		w.WriteHeader(http.StatusOK)
		return
	} else {
		http.Error(w, "Connection error", http.StatusInternalServerError)
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

func (us *URLShortener) JSONBatchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var requestBody []BatchRequestBody
	err = json.Unmarshal(body, &requestBody)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var responseBodyBatch []BatchResponseBodyItem
	var dataStorageRows []storage.DataStorageRow

	for _, requestBodyRow := range requestBody {
		shortKey, ok := us.getShortURL(requestBodyRow.OriginalURL)

		if !ok {
			shortKey = generateShortKey()
		}

		responseBody := BatchResponseBodyItem{
			CorrelationID: requestBodyRow.CorrelationID,
			ShortURL:      shortKey,
		}
		responseBodyBatch = append(responseBodyBatch, responseBody)

		dataStorageRow := storage.DataStorageRow{
			ShortURL: shortKey,
			URL:      requestBodyRow.OriginalURL,
		}
		dataStorageRows = append(dataStorageRows, dataStorageRow)

		if len(responseBodyBatch) == 1000 {
			us.saveBatch(w, dataStorageRows)
			dataStorageRows = dataStorageRows[:0]
		}
	}

	if len(dataStorageRows) > 0 {
		us.saveBatch(w, dataStorageRows)
	}

	err = us.buildJSONBatchResponse(w, responseBodyBatch)

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (us *URLShortener) saveBatch(w http.ResponseWriter, dataStorageRows []storage.DataStorageRow) {
	err := us.Storage.SaveBatch(dataStorageRows)

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
}

func (us *URLShortener) getShortKey(postURL string) (string, error) {
	shortKey, ok := us.getShortURL(postURL)

	if ok {
		return shortKey, nil
	}

	shortKey = generateShortKey()
	err := us.Storage.Save(shortKey, postURL)

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

func (us *URLShortener) buildJSONBatchResponse(w http.ResponseWriter, response []BatchResponseBodyItem) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

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
