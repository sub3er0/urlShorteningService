package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/sub3er0/urlShorteningService/internal/storage"
)

// MockURLStorage - структура-мок для интерфейса storage.URLStorageInterface
type MockURLStorage struct {
	mock.Mock
}

func (m *MockURLStorage) SetConnection(conn storage.DBConnectionInterface) {
	m.Called(conn)
}

// GetURL реализует метод интерфейса URLStorageInterface
func (m *MockURLStorage) GetURL(shortURL string) (storage.GetURLRow, bool) {
	args := m.Called(shortURL)
	return args.Get(0).(storage.GetURLRow), args.Bool(1)
}

// GetURLCount реализует метод интерфейса URLStorageInterface
func (m *MockURLStorage) GetURLCount() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

// GetShortURL реализует метод интерфейса URLStorageInterface
func (m *MockURLStorage) GetShortURL(URL string) (string, error) {
	args := m.Called(URL)
	return args.String(0), args.Error(1)
}

// Save реализует метод интерфейса URLStorageInterface
func (m *MockURLStorage) Save(ShortURL string, URL string, userID string) error {
	args := m.Called(ShortURL, URL, userID)
	return args.Error(0)
}

// LoadData реализует метод интерфейса URLStorageInterface
func (m *MockURLStorage) LoadData() ([]storage.DataStorageRow, error) {
	args := m.Called()
	return args.Get(0).([]storage.DataStorageRow), args.Error(1)
}

// Ping реализует метод интерфейса URLStorageInterface
func (m *MockURLStorage) Ping() bool {
	args := m.Called()
	return args.Bool(0)
}

// SaveBatch реализует метод интерфейса URLStorageInterface
func (m *MockURLStorage) SaveBatch(dataStorageRows []storage.DataStorageRow) error {
	args := m.Called(dataStorageRows)
	return args.Error(0)
}

// Init реализует метод интерфейса URLStorageInterface
func (m *MockURLStorage) Init(connectionString string) error {
	args := m.Called(connectionString)
	return args.Error(0)
}

// Close реализует метод интерфейса URLStorageInterface
func (m *MockURLStorage) Close() {
	m.Called()
}

// Тесты для URLRepository
func TestGetURL(t *testing.T) {
	mockStorage := new(MockURLStorage)
	repo := &URLRepository{Storage: mockStorage}

	// Подготовка ожидаемого результата
	expectedRow := storage.GetURLRow{URL: "http://example.com", IsDeleted: false}
	mockStorage.On("GetURL", "shorturl").Return(expectedRow, true)

	// Вызов метода GetURL
	row, found := repo.GetURL("shorturl")

	// Проверка результатов
	assert.True(t, found)
	assert.Equal(t, expectedRow, row)

	// Проверка, что ожидания выполнены
	mockStorage.AssertExpectations(t)
}

func TestGetURLCount(t *testing.T) {
	mockStorage := new(MockURLStorage)
	repo := &URLRepository{Storage: mockStorage}

	// Подготовка ожидания
	mockStorage.On("GetURLCount").Return(42, nil)

	// Вызов метода GetURLCount
	count, _ := repo.GetURLCount()

	// Проверка результата
	assert.Equal(t, 42, count)

	// Проверка ожиданий
	mockStorage.AssertExpectations(t)
}

func TestGetShortURL(t *testing.T) {
	mockStorage := new(MockURLStorage)
	repo := &URLRepository{Storage: mockStorage}

	// Подготовка ожидаемого результата
	mockStorage.On("GetShortURL", "http://example.com").Return("shorturl", nil)

	// Вызов метода GetShortURL
	shortURL, err := repo.GetShortURL("http://example.com")

	// Проверка результата
	assert.NoError(t, err)
	assert.Equal(t, "shorturl", shortURL)

	// Проверка ожиданий
	mockStorage.AssertExpectations(t)
}

func TestSave(t *testing.T) {
	mockStorage := new(MockURLStorage)
	repo := &URLRepository{Storage: mockStorage}

	// Подготовка ожидания
	mockStorage.On("Save", "shorturl", "http://example.com", "user123").Return(nil)

	// Вызов метода Save
	err := repo.Save("shorturl", "http://example.com", "user123")

	// Проверка ошибок
	assert.NoError(t, err)

	// Проверка ожиданий
	mockStorage.AssertExpectations(t)
}

func TestLoadData(t *testing.T) {
	mockStorage := new(MockURLStorage)
	repo := &URLRepository{Storage: mockStorage}

	// Подготовка ожидаемого результата
	expectedRows := []storage.DataStorageRow{{URL: "http://example.com", ShortURL: "shorturl"}}
	mockStorage.On("LoadData").Return(expectedRows, nil)

	// Вызов метода LoadData
	data, err := repo.LoadData()

	// Проверка ошибок
	assert.NoError(t, err)
	assert.Equal(t, expectedRows, data)

	// Проверка ожиданий
	mockStorage.AssertExpectations(t)
}

func TestPing(t *testing.T) {
	mockStorage := new(MockURLStorage)
	repo := &URLRepository{Storage: mockStorage}

	// Подготовка ожидания
	mockStorage.On("Ping").Return(true)

	// Вызов метода Ping
	result := repo.Ping()

	// Проверка результата
	assert.True(t, result)

	// Проверка ожиданий
	mockStorage.AssertExpectations(t)
}

func TestSaveBatch(t *testing.T) {
	mockStorage := new(MockURLStorage)
	repo := &URLRepository{Storage: mockStorage}

	// Подготовка данных для сохранения
	dataStorageRows := []storage.DataStorageRow{
		{URL: "http://example.com", ShortURL: "shorturl"},
		{URL: "http://example.org", ShortURL: "shorturl2"},
	}

	// Установка ожидания
	mockStorage.On("SaveBatch", dataStorageRows).Return(nil)

	// Вызов метода SaveBatch
	err := repo.SaveBatch(dataStorageRows)

	// Проверка ошибок
	assert.NoError(t, err)

	// Проверка ожиданий
	mockStorage.AssertExpectations(t)
}
