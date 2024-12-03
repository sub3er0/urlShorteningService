package repository_test

import (
	"github.com/stretchr/testify/mock"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/sub3er0/urlShorteningService/internal/repository"
	"github.com/sub3er0/urlShorteningService/internal/storage"
)

// MockUserStorage - структура, реализующая интерфейс UserStorageInterface.
type MockUserStorage struct {
	mock.Mock
}

// IsUserExist реализует метод интерфейса UserStorageInterface.
func (m *MockUserStorage) IsUserExist(uniqueID string) bool {
	args := m.Called(uniqueID)
	return args.Bool(0)
}

// SaveUser реализует метод интерфейса UserStorageInterface.
func (m *MockUserStorage) SaveUser(uniqueID string) error {
	args := m.Called(uniqueID)
	return args.Error(0)
}

// GetUserUrls реализует метод интерфейса UserStorageInterface.
func (m *MockUserStorage) GetUserUrls(uniqueID string) ([]storage.UserUrlsResponseBodyItem, error) {
	args := m.Called(uniqueID)
	return args.Get(0).([]storage.UserUrlsResponseBodyItem), args.Error(1)
}

// DeleteUserUrls реализует метод интерфейса UserStorageInterface.
func (m *MockUserStorage) DeleteUserUrls(uniqueID string, shortURLs []string) error {
	args := m.Called(uniqueID, shortURLs)
	return args.Error(0)
}

// Init реализует метод интерфейса UserStorageInterface.
func (m *MockUserStorage) Init(connectionString string) error {
	args := m.Called(connectionString)
	return args.Error(0)
}

// Close реализует метод интерфейса UserStorageInterface.
func (m *MockUserStorage) Close() {
	m.Called()
}

func (m *MockUserStorage) GetUsersCount() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

func TestIsUserExist(t *testing.T) {
	mockStorage := new(MockUserStorage)
	repo := &repository.UserRepository{Storage: mockStorage}

	mockStorage.On("IsUserExist", "user123").Return(true)

	exists := repo.IsUserExist("user123")

	assert.True(t, exists)
	mockStorage.AssertExpectations(t)
}

func TestSaveUser(t *testing.T) {
	mockStorage := new(MockUserStorage)
	repo := &repository.UserRepository{Storage: mockStorage}

	mockStorage.On("SaveUser", "user123").Return(nil)

	err := repo.SaveUser("user123")

	assert.NoError(t, err)
	mockStorage.AssertExpectations(t)
}

func TestGetUserUrls(t *testing.T) {
	mockStorage := new(MockUserStorage)
	repo := &repository.UserRepository{Storage: mockStorage}

	expectedUrls := []storage.UserUrlsResponseBodyItem{{OriginalURL: "http://example.com", ShortURL: "shorturl"}}
	mockStorage.On("GetUserUrls", "user123").Return(expectedUrls, nil)

	urls, err := repo.GetUserUrls("user123")

	assert.NoError(t, err)
	assert.Equal(t, expectedUrls, urls)
	mockStorage.AssertExpectations(t)
}

func TestDeleteUserUrls(t *testing.T) {
	mockStorage := new(MockUserStorage)
	repo := &repository.UserRepository{Storage: mockStorage}

	mockStorage.On("DeleteUserUrls", "user123", []string{"shorturl"}).Return(nil)

	err := repo.DeleteUserUrls("user123", []string{"shorturl"})

	assert.NoError(t, err)
	mockStorage.AssertExpectations(t)
}
