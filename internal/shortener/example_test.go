package shortener

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
)

// ExamplePostHandler демонстрирует, как использовать PostHandler.
func ExampleURLShortener_PostHandler() {
	us := &URLShortener{
		BaseURL: "http://short.url/",
	}

	body := `{"url": "http://example.com"}`
	req := httptest.NewRequest("POST", "/", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	us.PostHandler(w, req) // Вызов метода для создания короткого URL.

	res := w.Result()
	jsonResponse, _ := json.Marshal(res) // Преобразование результата в JSON.

	// В этом примере нам нужно проверить содержимое jsonResponse
	// или статус через различные ассерты.
	fmt.Println(string(jsonResponse)) // Выводим результат ответа в консоль.
}

// ExampleGetHandler демонстрирует использование GetHandler.
func ExampleURLShortener_GetHandler() {
	us := &URLShortener{}

	req := httptest.NewRequest("GET", "/url/someShortURL", nil)
	w := httptest.NewRecorder()

	us.GetHandler(w, req)

	res := w.Result()
	fmt.Printf("URL Status: %d", res.StatusCode) // Печатаем статус ответа
}

// ExampleJSONPostHandler демонстрирует использование JSONPostHandler.
func ExampleURLShortener_JSONPostHandler() {
	us := &URLShortener{
		BaseURL: "http://short.url/",
	}

	body := `{"url": "http://example.com"}`
	req := httptest.NewRequest("POST", "/shorten", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	us.JSONPostHandler(w, req)

	res := w.Result()
	fmt.Printf("Response: %v", res)
}

// ExamplePingHandler демонстрирует использование метода PingHandler.
func ExampleURLShortener_PingHandler() {
	us := &URLShortener{}

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	us.PingHandler(w, req)

	res := w.Result()
	println("Ping Response Status:", res.Status)
}

// ExampleJSONBatchHandler демонстрирует использование метода JSONBatchHandler.
func ExampleURLShortener_JSONBatchHandler() {
	us := &URLShortener{
		BaseURL: "http://short.url/",
	}

	// Подготовка данных для пакетного запроса
	requestBody := []BatchRequestBody{
		{CorrelationID: "12345", OriginalURL: "http://example.com"},
	}
	jsonBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/shorten/batch", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	us.JSONBatchHandler(w, req)

	res := w.Result()
	println("Batch Response Status:", res.Status)
}
