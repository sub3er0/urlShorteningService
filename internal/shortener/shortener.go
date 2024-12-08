package shortener

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"sync"

	"github.com/pkg/errors"
	"github.com/sub3er0/urlShorteningService/internal/cookie"
	"github.com/sub3er0/urlShorteningService/internal/repository"
	"github.com/sub3er0/urlShorteningService/internal/storage"
)

// URLShortener представляет структуру, ответственную за обработку
// запросов на создание коротких URL и управление взаимодействиями с
// хранилищами URL и пользователей.
type URLShortener struct {
	// URLRepository предоставляет доступ к операциям работы с URL в хранилище.
	URLRepository repository.URLRepositoryInterface

	// UserRepository предоставляет доступ к операциям работы с пользователями в хранилище.
	UserRepository repository.UserRepositoryInterface

	// ServerAddress определяет адрес HTTP-сервера, на котором будет работать приложение.
	ServerAddress string

	// BaseURL представляет базовый адрес, который используется для сокращённых URL.
	BaseURL string

	// TrustedSubnet доверенная подсеть
	TrustedSubnet string

	// CookieManager управляет аутентификацией и обработкой куки в приложении.
	CookieManager cookie.CookieManagerInterface

	// RemoveChan — это канал, который используется для передачи коротких URL, которые нужно удалить.
	RemoveChan chan string

	// wg используется для управления ожидающими горутинами.
	wg sync.WaitGroup
}

// URLShortenerInterface - интерфейс для работы с сокращениями URL.
type URLShortenerInterface interface {
	// GetHandler Получает короткий URL из репозитория
	GetHandler(w http.ResponseWriter, r *http.Request)

	// PingHandler Проверяет состояние соединения с репозиторием
	PingHandler(w http.ResponseWriter, r *http.Request)

	// JSONPostHandler Обрабатывает запрос на создание короткого URL в формате JSON
	JSONPostHandler(w http.ResponseWriter, r *http.Request)

	// PostHandler Обрабатывает запрос на создание короткого URL
	PostHandler(w http.ResponseWriter, r *http.Request)

	// JSONBatchHandler Обрабатывает пакетные запросы на создание сокращенных URL
	JSONBatchHandler(w http.ResponseWriter, r *http.Request)

	// GetUserUrls Получает URL пользователя
	GetUserUrls(w http.ResponseWriter, r *http.Request)

	// DeleteUserUrls Удаляет короткие URL
	DeleteUserUrls(w http.ResponseWriter, r *http.Request)

	// Worker Удаляет короткие URL
	Worker()

	// IsIPInTrustedSubnet проверка доверенной подсети
	IsIPInTrustedSubnet(ip string) bool
}

// JSONResponseBody представляет структуру для ответа в формате JSON.
// Включает поле Result, содержащее результат выполнения какой-либо операции.
type JSONResponseBody struct {
	Result string `json:"result"` // Результат выполнения, представляет собой строку.
}

// RequestBody представляет структуру для запроса, содержащего URL.
// Используется при получении короткого URL.
type RequestBody struct {
	URL string `json:"url"` // Полный URL для сокращения.
}

// DeleteRequestBody представляет структуру для запроса на удаление короткого URL.
// Служит для передачи данных, необходимых для операций удаления.
type DeleteRequestBody struct {
	ShortURL string `json:"short_url"` // Короткий URL, который необходимо удалить.
}

// BatchRequestBody представляет структуру для пакетных запросов на создание сокращенных URL.
// Содержит идентификатор корреляции и оригинальный URL.
type BatchRequestBody struct {
	CorrelationID string `json:"correlation_id"` // Идентификатор корреляции для отслеживания в запросах.
	OriginalURL   string `json:"original_url"`   // Оригинальный URL, который будет сокращён.
}

// BatchResponseBodyItem представляет элемент ответа для пакетных операций по сокращению URL.
// Содержит идентификатор корреляции и сокращенный URL.
type BatchResponseBodyItem struct {
	CorrelationID string `json:"correlation_id"` // Идентификатор корреляции для сопоставления с запросом.
	ShortURL      string `json:"short_url"`      // Сокращенный URL.
}

// ExistValueError представляет пользовательскую ошибку для случаев,
// когда значение уже существует в системе.
type ExistValueError struct {
	Text string // Сообщение об ошибке.
}

// ErrShortURLExists указывает на ошибку, возникающую при попытке сохранить
// короткий URL, который уже существует в хранилище.
var ErrShortURLExists = &ExistValueError{Text: "ShortURL already exists"}

// Worker Удаляет короткие URL
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

// Error возвращает текст сообщения об ошибке в формате строки.
func (e *ExistValueError) Error() string {
	return e.Text
}

// getShortURL возвращает короткий URL для заданного полного URL.
// Если в репозитории не найдено, возвращает ошибку.
// Параметры:
//   - URL: полный URL для получения короткого URL.
//
// Возвращает короткий URL и ошибку, если произошла проблема.
func (us *URLShortener) getShortURL(URL string) (string, error) {
	return us.URLRepository.GetShortURL(URL)
}

// GetHandler Получает короткий URL из репозитория
func (us *URLShortener) GetHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	storedURL, ok := us.URLRepository.GetURL(id)

	if !ok {
		http.Error(w, "NotFound", http.StatusNotFound)
	} else if !storedURL.IsDeleted {
		w.Header().Set("Location", storedURL.URL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else if storedURL.IsDeleted {
		w.WriteHeader(http.StatusGone)
	}
}

// GetInternalStats Получение статистики сервиса
func (us *URLShortener) GetInternalStats(w http.ResponseWriter, r *http.Request) {
	clientIP := r.Header.Get("X-Real-IP")
	if !us.IsIPInTrustedSubnet(clientIP) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	urlsCount, err := us.URLRepository.GetURLCount()

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	usersCount, err := us.UserRepository.GetUsersCount()

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Формируем ответ
	stats := map[string]int{
		"urls":  urlsCount,
		"users": usersCount,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// проверка, входит ли IP-адрес в доверенную подсеть
func (us *URLShortener) IsIPInTrustedSubnet(ip string) bool {
	if us.TrustedSubnet == "" {
		return false
	}

	_, subnet, err := net.ParseCIDR(us.TrustedSubnet)
	if err != nil {
		return false
	}

	clientIP := net.ParseIP(ip)
	if clientIP == nil {
		return false
	}

	return subnet.Contains(clientIP)
}

// PingHandler Проверяет состояние соединения с репозиторием
func (us *URLShortener) PingHandler(w http.ResponseWriter, r *http.Request) {
	ok := us.URLRepository.Ping()

	if ok {
		w.WriteHeader(http.StatusOK)
		return
	} else {
		http.Error(w, "Connection error", http.StatusInternalServerError)
	}
}

// JSONPostHandler Обрабатывает запрос на создание короткого URL в формате JSON
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

// JSONBatchHandler Обрабатывает пакетные запросы на создание сокращенных URL
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
		shortKey, getShortURLError := us.getShortURL(requestBodyRow.OriginalURL)

		if getShortURLError != nil {
			shortKey = GenerateShortKey()
		}

		responseBody := BatchResponseBodyItem{
			CorrelationID: requestBodyRow.CorrelationID,
			ShortURL:      us.BaseURL + shortKey,
		}

		if errors.Is(getShortURLError, ErrShortURLExists) {
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
			getShortURLError = us.URLRepository.SaveBatch(dataStorageRows)
			log.Printf("ERROR = %v", getShortURLError)

			if getShortURLError != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			dataStorageRows = dataStorageRows[:0]
			us.URLRepository.Save(shortKey, requestBodyRow.OriginalURL, us.CookieManager.GetActualCookieValue())
		}
	}

	if len(dataStorageRows) > 0 {
		err = us.URLRepository.SaveBatch(dataStorageRows)
		log.Printf("ERROR = %v", err)

		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	err = us.buildJSONBatchResponse(w, responseBodyBatch)

	if err != nil {
		log.Printf("Internal Server Error")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	log.Printf("END BATCH HADLER")
}

// GetUserUrls Получает URL пользователя
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

func (us *URLShortener) saveBatch(w http.ResponseWriter, dataStorageRows []storage.DataStorageRow) error {
	err := us.URLRepository.SaveBatch(dataStorageRows)

	if err != nil {
		return fmt.Errorf("ServerAddress is required")
	}

	return nil
}

// PostHandler Обрабатывает запрос на создание короткого URL
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

// getShortKey генерирует короткий ключ для заданного оригинального URL.
// Если короткий URL уже существует, возвращает его и ошибку ErrShortURLExists.
// Если короткого URL не существует, он создается и сохраняется в репозитории.
// Параметры:
//   - postURL: оригинальный URL, для которого требуется получить или создать короткий ключ.
//
// Возвращает короткий ключ и ошибку, если возникла проблема.
func (us *URLShortener) getShortKey(postURL string) (string, error) {
	shortKey, err := us.URLRepository.GetShortURL(postURL)

	if err == nil {
		return shortKey, ErrShortURLExists
	}

	shortKey = GenerateShortKey()
	err = us.URLRepository.Save(shortKey, postURL, us.CookieManager.GetActualCookieValue())

	if err != nil {
		return "", err
	}

	return shortKey, nil
}

// GenerateShortKey создает новый короткий ключ длиной 6 знаков, состоящий из
// букв и цифр. Использует криптографически безопасный генератор случайных чисел.
// Возвращает сгенерированный короткий ключ.
func GenerateShortKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLength = 6

	shortKey := make([]byte, keyLength)

	for i := range shortKey {
		shortKey[i] = charset[rand.Intn(len(charset))]
	}

	return string(shortKey)
}

// buildResponse формирует ответ на запрос с коротким URL.
// Устанавливает заголовок типа контента и статус ответа в зависимости от того,
// существует ли короткий URL или нет
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

// buildJSONResponse формирует JSON-ответ для HTTP-запроса с указанным результатом.
// Устанавливает заголовок "Content-Type" в "application/json" и устанавливает статус ответа
// в зависимости от существования короткого URL.
// Параметры:
//   - w: объект ResponseWriter для записи ответа.
//   - response: данные для сериализации в формате JSON.
//   - isExist: булево значение, указывающее, существует ли короткий URL.
//
// Возвращает ошибку, если произошла проблема с сериализацией или записью ответа.
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

// buildJSONBatchResponse формирует JSON-ответ для пакетных запросов с указанными данными.
// Устанавливает заголовок "Content-Type" в "application/json" и возвращает статус 201 Created.
// Параметры:
//   - w: объект ResponseWriter для записи ответа.
//   - response: массив данных для сериализации в формате JSON.
//
// Возвращает ошибку, если произошла проблема с сериализацией или записью ответа.
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

// buildAllUserUrlsResponse формирует JSON-ответ для списка URL пользователя.
// Устанавливает заголовок "Content-Type" в "application/json" и статус ответа в 200 OK.
// Параметры:
//   - w: объект ResponseWriter для записи ответа.
//   - response: массив строк данных URL пользователя, которые будут сериализованы в формате JSON.
//
// Возвращает ошибку, если произошла проблема с сериализацией данных или записью ответа.
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

// DeleteUserUrlsBatch удаляет пакетные короткие URL для текущего пользователя.
// Принимает массив коротких URL и обрабатывает их удаление в партиях заданного размера.
//
// Параметры:
//   - shortURLs: массив коротких URL, которые необходимо удалить.
//
// Метод не возвращает значений. Если возникает ошибка при удалении любого из URL,
// она будет записана в лог, но выполнение продолжится для следующих URL.
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

// DeleteUserUrls Удаляет короткие URL
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
