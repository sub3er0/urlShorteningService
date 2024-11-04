package shortener

import (
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sync"

	"github.com/pkg/errors"
	"github.com/sub3er0/urlShorteningService/internal/cookie"
	"github.com/sub3er0/urlShorteningService/internal/repository"
	"github.com/sub3er0/urlShorteningService/internal/storage"
)

// URLShortener Структура URLShortener, использующая интерфейс хранения
type URLShortener struct {
	UrlRepository  repository.UrlRepositoryInterface
	UserRepository repository.UserRepositoryInterface
	ServerAddress  string
	BaseURL        string
	CookieManager  cookie.CookieManagerInterface
	RemoveChan     chan string
	wg             sync.WaitGroup
}

// URLShortenerInterface - интерфейс для работы с сокращениями URL.
type URLShortenerInterface interface {
	// GetHandler Получает короткий URL из репозитория
	GetHandler(w http.ResponseWriter, r *http.Request)

	// PingHandler Проверяет состояние соединения с репозиторием
	PingHandler(w http.ResponseWriter, r *http.Request)

	// JSONPostHandler Обрабатывает запрос на создание короткого URL в формате JSON
	JSONPostHandler(w http.ResponseWriter, r *http.Request)

	// JSONBatchHandler Обрабатывает пакетные запросы на создание сокращенных URL
	JSONBatchHandler(w http.ResponseWriter, r *http.Request)

	// GetUserUrls Получает URL пользователя
	GetUserUrls(w http.ResponseWriter, r *http.Request)

	// DeleteUserUrls Удаляет короткие URL
	DeleteUserUrls(w http.ResponseWriter, r *http.Request)

	// Worker Удаляет короткие URL
	Worker()
}

type JSONResponseBody struct {
	Result string `json:"result"`
}

type RequestBody struct {
	URL string `json:"url"`
}

type DeleteRequestBody struct {
	ShortURL string `json:"short_url"`
}

type BatchRequestBody struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchResponseBodyItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type ExistValueError struct {
	Text string
}

var ErrShortURLExists = &ExistValueError{Text: "ShortURL already exists"}

func (us *URLShortener) Worker() {
	batchSize := 1
	shortURLs := make([]string, 0, batchSize)

	for urlFromChan := range us.RemoveChan {
		shortURLs = append(shortURLs, urlFromChan)

		if len(shortURLs) >= batchSize {
			err := us.UserRepository.DeleteUserUrls(us.CookieManager.GetActualCookieValue(), shortURLs)
			if err != nil {
				log.Printf("Error while deleting urls")
			}
			shortURLs = shortURLs[:0]
		}
	}

	if len(shortURLs) > 0 {
		if err := us.UserRepository.DeleteUserUrls(us.CookieManager.GetActualCookieValue(), shortURLs); err != nil {
			log.Printf("Error while deleting remaining URLs: %v", err)
		}
	}
}

func (e *ExistValueError) Error() string {
	return e.Text
}

func (us *URLShortener) getShortURL(URL string) (string, error) {
	return us.UrlRepository.GetShortURL(URL)
}

func (us *URLShortener) GetHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	storedURL, ok := us.UrlRepository.GetURL(id)

	if !ok {
		http.Error(w, "NotFound", http.StatusNotFound)
	} else if !storedURL.IsDeleted {
		w.Header().Set("Location", storedURL.URL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else if storedURL.IsDeleted {
		w.WriteHeader(http.StatusGone)
	}

	return
}

func (us *URLShortener) PingHandler(w http.ResponseWriter, r *http.Request) {
	ok := us.UrlRepository.Ping()

	if ok {
		w.WriteHeader(http.StatusOK)
		return
	} else {
		http.Error(w, "Connection error", http.StatusInternalServerError)
	}
}

func (us *URLShortener) JSONPostHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var requestBody RequestBody
	err = json.Unmarshal(body, &requestBody)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	bodyURL, err := url.ParseRequestURI(requestBody.URL)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shortKey, err := us.getShortKey(bodyURL.String())

	var responseBody JSONResponseBody
	responseBody.Result = shortKey

	if errors.Is(err, ErrShortURLExists) {
		err = us.buildJSONResponse(w, responseBody, true)
	} else if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	} else {
		err = us.buildJSONResponse(w, responseBody, false)
	}

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (us *URLShortener) JSONBatchHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var requestBody []BatchRequestBody
	err = json.Unmarshal(body, &requestBody)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var responseBodyBatch []BatchResponseBodyItem
	var dataStorageRows []storage.DataStorageRow

	for _, requestBodyRow := range requestBody {
		shortKey, err := us.getShortURL(requestBodyRow.OriginalURL)

		if err != nil {
			shortKey = generateShortKey()
		}

		responseBody := BatchResponseBodyItem{
			CorrelationID: requestBodyRow.CorrelationID,
			ShortURL:      us.BaseURL + shortKey,
		}

		if errors.Is(err, ErrShortURLExists) {
			responseBodyBatch = append(responseBodyBatch, responseBody)
			continue
		}

		responseBodyBatch = append(responseBodyBatch, responseBody)

		dataStorageRow := storage.DataStorageRow{
			ShortURL: shortKey,
			URL:      requestBodyRow.OriginalURL,
			UserID:   us.CookieManager.GetActualCookieValue(),
		}
		dataStorageRows = append(dataStorageRows, dataStorageRow)

		if len(responseBodyBatch) == 1000 {
			us.saveBatch(w, dataStorageRows)
			dataStorageRows = dataStorageRows[:0]
			us.UrlRepository.Save(shortKey, requestBodyRow.OriginalURL, us.CookieManager.GetActualCookieValue())
		}
	}

	if len(dataStorageRows) > 0 {
		us.saveBatch(w, dataStorageRows)
	}

	err = us.buildJSONBatchResponse(w, responseBodyBatch)

	if err != nil {
		log.Printf("Internal Server Error")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	log.Printf("END BATCH HADLER")
}

func (us *URLShortener) GetUserUrls(w http.ResponseWriter, r *http.Request) {
	urls, err := us.UserRepository.GetUserUrls(us.CookieManager.GetActualCookieValue())

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if len(urls) == 0 {
		http.Error(w, "No Content", http.StatusNoContent)
		return
	}

	err = us.buildAllUserUrlsResponse(w, urls)

	if err != nil {
		return
	}
}

func (us *URLShortener) saveBatch(w http.ResponseWriter, dataStorageRows []storage.DataStorageRow) {
	err := us.UrlRepository.SaveBatch(dataStorageRows)

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (us *URLShortener) PostHandler(w http.ResponseWriter, r *http.Request) {
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

	if errors.Is(err, ErrShortURLExists) {
		us.buildResponse(w, shortKey, true)
	} else if err != nil {
		http.Error(w, "Internal Server Error", http.StatusBadRequest)
		return
	} else {
		us.buildResponse(w, shortKey, false)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()

		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}(r.Body)
}

func (us *URLShortener) getShortKey(postURL string) (string, error) {
	shortKey, err := us.UrlRepository.GetShortURL(postURL)

	if err == nil {
		return shortKey, ErrShortURLExists
	}

	shortKey = generateShortKey()
	err = us.UrlRepository.Save(shortKey, postURL, us.CookieManager.GetActualCookieValue())

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

func (us *URLShortener) buildResponse(w http.ResponseWriter, shortKey string, isExist bool) {
	w.Header().Set("content-type", "text/plain")

	if !isExist {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusConflict)
	}

	if len(us.BaseURL) > 0 && us.BaseURL[len(us.BaseURL)-1] != '/' {
		us.BaseURL = us.BaseURL + "/"
	}

	_, err := w.Write([]byte(us.BaseURL + shortKey))

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (us *URLShortener) buildJSONResponse(w http.ResponseWriter, response JSONResponseBody, isExist bool) error {
	w.Header().Set("Content-Type", "application/json")
	if !isExist {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusConflict)
	}

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

func (us *URLShortener) buildAllUserUrlsResponse(w http.ResponseWriter, response []storage.UserUrlsResponseBodyItem) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	for i := range response {
		response[i].ShortURL = us.BaseURL + response[i].ShortURL
	}

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

func (us *URLShortener) DeleteUserUrlsBatch(shortURLs []string) {
	batchSize := 100

	for i := 0; i < len(shortURLs); i += batchSize {
		end := i + batchSize
		if end > len(shortURLs) {
			end = len(shortURLs)
		}

		urlsBatch := shortURLs[i:end]
		err := us.UserRepository.DeleteUserUrls(us.CookieManager.GetActualCookieValue(), urlsBatch)
		if err != nil {
			log.Printf("Error while deleting urls")
		}
	}
}

func (us *URLShortener) DeleteUserUrls(w http.ResponseWriter, r *http.Request) {
	var shortURLs []string
	err := json.NewDecoder(r.Body).Decode(&shortURLs)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, shortURL := range shortURLs {
		us.RemoveChan <- shortURL
	}

	w.WriteHeader(http.StatusAccepted)
}
