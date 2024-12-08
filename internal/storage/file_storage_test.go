package storage

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

// Путь к тестовому файлу
const testFilePath = "test_data.json"

// Очищает тестовый файл перед каждым тестом
func clearTestFile() {
	os.Remove(testFilePath)
}

// Тест для метода Save
func TestSave(t *testing.T) {
	clearTestFile()
	defer clearTestFile()

	fs := &FileStorage{FileStoragePath: testFilePath}
	err := fs.Save("short1", "http://example.com", "user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rows, err := fs.LoadData()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rows) != 1 || rows[0].ShortURL != "short1" || rows[0].URL != "http://example.com" {
		t.Errorf("unexpected data in storage: %+v", rows)
	}
}

// Тест для метода GetURL
func TestGetURL(t *testing.T) {
	clearTestFile()
	defer clearTestFile()

	fs := &FileStorage{FileStoragePath: testFilePath}
	fs.Save("short1", "http://example.com", "user1")

	getURLRow, found := fs.GetURL("short1")
	if !found {
		t.Fatal("expected to find short URL")
	}
	if getURLRow.URL != "http://example.com" {
		t.Errorf("expected URL to be 'http://example.com', got '%s'", getURLRow.URL)
	}
}

// Тест для метода GetShortURL
func TestGetShortURL(t *testing.T) {
	clearTestFile()
	defer clearTestFile()

	fs := &FileStorage{FileStoragePath: testFilePath}
	err := fs.Save("short1", "http://example.com", "user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	shortURL, err := fs.GetShortURL("http://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if shortURL != "short1" {
		t.Errorf("expected short URL to be 'short1', got '%s'", shortURL)
	}
}

// Тест для метода LoadData
func TestLoadData(t *testing.T) {
	clearTestFile()
	defer clearTestFile()

	fs := &FileStorage{FileStoragePath: testFilePath}
	_ = fs.Save("short1", "http://example.com", "user1")
	_ = fs.Save("short2", "http://example.org", "user2")

	dataRows, err := fs.LoadData()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(dataRows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(dataRows))
	}
}

// Тест для метода GetURLCount
func TestGetURLCount(t *testing.T) {
	clearTestFile()
	defer clearTestFile()

	fs := &FileStorage{FileStoragePath: testFilePath}

	count, _ := fs.GetURLCount()

	if count != 0 {
		t.Errorf("expected URL count to be 0, got %d", count)
	}

	_ = fs.Save("short1", "http://example.com", "user1")

	count, _ = fs.GetURLCount()

	if count != 1 {
		t.Errorf("expected URL count to be 1, got %d", count)
	}
}

func TestSaveBatch(t *testing.T) {
	testFilePath := "test_file_storage.json"
	defer os.Remove(testFilePath) // Удаляем файл после теста

	fs := &FileStorage{FileStoragePath: testFilePath}

	dataRows := []DataStorageRow{
		{ID: 1, ShortURL: "short.ly/1", URL: "http://example.com", UserID: "user1", DeletedFlag: false},
		{ID: 2, ShortURL: "short.ly/2", URL: "http://example.org", UserID: "user2", DeletedFlag: false},
	}

	err := fs.SaveBatch(dataRows)
	assert.NoError(t, err)

	content, err := os.ReadFile(testFilePath)
	assert.NoError(t, err)

	var savedRows []DataStorageRow
	err = json.Unmarshal(content, &savedRows)

	assert.NoError(t, err)
	assert.Len(t, savedRows, 2)
}

func TestPing(t *testing.T) {
	fs := &FileStorage{}
	assert.True(t, fs.Ping())
}

func TestIsUserExist(t *testing.T) {
	fs := &FileStorage{}
	assert.False(t, fs.IsUserExist("test_user_id"))
}

func TestSaveUser(t *testing.T) {
	fs := &FileStorage{}
	err := fs.SaveUser("test_user_id")
	assert.NoError(t, err)
}

func TestGetUserUrls(t *testing.T) {
	fs := &FileStorage{}
	_, err := fs.GetUserUrls("test_user_id")
	assert.NoError(t, err)
}

func TestDeleteUserUrls(t *testing.T) {
	fs := &FileStorage{}
	err := fs.DeleteUserUrls("unique_id", []string{"s", "s2"})
	assert.NoError(t, err)
}

func TestGetUsersCount(t *testing.T) {
	fs := &FileStorage{}
	_, err := fs.GetUsersCount()
	assert.NoError(t, err)
}
