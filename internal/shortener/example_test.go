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

	us.PostHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	var responseBody JSONResponseBody
	if err := json.NewDecoder(res.Body).Decode(&responseBody); err == nil {
		fmt.Println("Response:", responseBody.Result)
	} else {
		fmt.Println("Error decoding response:", err)
	}
}

// ExampleGetHandler демонстрирует использование GetHandler.
func ExampleURLShortener_GetHandler() {
	us := &URLShortener{}

	req := httptest.NewRequest("GET", "/url/someShortURL", nil)
	w := httptest.NewRecorder()

	us.GetHandler(w, req)

	res := w.Result()
	defer res.Body.Close()
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
	defer res.Body.Close()
	fmt.Printf("Response: %v", res)
}

// ExamplePingHandler демонстрирует использование метода PingHandler.
func ExampleURLShortener_PingHandler() {
	us := &URLShortener{}

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	us.PingHandler(w, req)

	res := w.Result()
	defer res.Body.Close()
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
	defer res.Body.Close()
	println("Batch Response Status:", res.Status)
}
