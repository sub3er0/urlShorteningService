package repository

import "github.com/sub3er0/urlShorteningService/internal/storage"

type UserRepositoryInterface interface {
	IsUserExist(uniqueID string) bool
	SaveUser(uniqueID string) error
	GetUserUrls(uniqueID string) ([]storage.UserUrlsResponseBodyItem, error)
	DeleteUserUrls(uniqueID string, shortURLS []string) error
}

// UserRepository реализует UserRepositoryInterface.
type UserRepository struct {
	Storage storage.UserStorageInterface
}

// IsUserExist проверяет, существует ли пользователь по уникальному ID.
func (ur *UserRepository) IsUserExist(uniqueID string) bool {
	return false
}

// SaveUser сохраняет пользователя с указанным уникальным ID.
func (ur *UserRepository) SaveUser(uniqueID string) error {
	return nil
}

// GetUserUrls возвращает список URL-адресов для указанного уникального ID пользователя.
func (ur *UserRepository) GetUserUrls(uniqueID string) ([]storage.UserUrlsResponseBodyItem, error) {
	return []storage.UserUrlsResponseBodyItem{}, nil
}

// DeleteUserUrls удаляет указанные URL-адреса для указанного уникального ID пользователя.
func (ur *UserRepository) DeleteUserUrls(uniqueID string, shortURLS []string) error {
	return nil
}
