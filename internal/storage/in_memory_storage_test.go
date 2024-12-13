package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInMemoryStorage(t *testing.T) {
	storage := &InMemoryStorage{Urls: make(map[string]string)}

	// Тестируем Init
	err := storage.Init("")
	assert.NoError(t, err, "Init should not return an error")

	// Тестируем Save
	shortURL := "short.ly/xyz"
	url := "http://example.com"
	userID := "user123"

	err = storage.Save(shortURL, url, userID)
	assert.NoError(t, err, "Save should not return an error")

	// Проверяем, что URL сохранен
	assert.Equal(t, url, storage.Urls[shortURL], "Saved URL should match")

	// Тестим GetURL
	getURLRow, ok := storage.GetURL(shortURL)
	assert.True(t, ok, "GetURL should return true")
	assert.Equal(t, url, getURLRow.URL, "GetURL should return the correct URL")

	// Тестим GetShortURL
	retrievedShortURL, err := storage.GetShortURL(url)
	assert.NoError(t, err, "GetShortURL should not return an error")
	assert.Equal(t, shortURL, retrievedShortURL, "GetShortURL should return the correct short URL")

	// Проверка на несуществующий URL
	_, err = storage.GetShortURL("http://nonexistent.com")
	assert.Error(t, err, "Expected error for nonexistent URL")

	// Тестим GetURLCount
	count, _ := storage.GetURLCount()
	assert.Equal(t, 1, count, "Expected URL count to be 1")

	// Тестим SaveBatch
	dataRows := []DataStorageRow{
		{ShortURL: "short.ly/abc", URL: "http://example2.com"},
		{ShortURL: "short.ly/def", URL: "http://example3.com"},
	}
	err = storage.SaveBatch(dataRows)
	assert.NoError(t, err, "SaveBatch should not return an error")

	// Проверяем сохранение нескольких URL
	for _, row := range dataRows {
		assert.Equal(t, row.URL, storage.Urls[row.ShortURL], "Expected saved URL to match")
	}

	// Тестим LoadData
	loadedData, err := storage.LoadData()
	assert.NoError(t, err, "LoadData should not return an error")
	assert.Empty(t, loadedData, "Expected loaded data length to be 0")

	// Тестирование Ping
	assert.True(t, storage.Ping(), "Expected Ping to return true")

	_ = storage.Set("shortURL", "longURL")
	url, _ = storage.GetShortURL("longURL")
	assert.Equal(t, "shortURL", url)
}
