package storage

import (
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

	if count := fs.GetURLCount(); count != 0 {
		t.Errorf("expected URL count to be 0, got %d", count)
	}

	_ = fs.Save("short1", "http://example.com", "user1")

	if count := fs.GetURLCount(); count != 1 {
		t.Errorf("expected URL count to be 1, got %d", count)
	}
}
