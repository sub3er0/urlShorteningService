package repository

import "github.com/sub3er0/urlShorteningService/internal/storage"

// UserRepositoryInterface определяет методы для работы с репозиторием пользователей.
// Этот интерфейс предоставляет доступ к операциям проверки существования пользователя,
// сохранения пользователей, получения и удаления их URL.
type UserRepositoryInterface interface {
	// IsUserExist проверяет, существует ли пользователь по его уникальному идентификатору.
	IsUserExist(uniqueID string) bool

	// SaveUser сохраняет нового пользователя с указанным уникальным идентификатором.
	SaveUser(uniqueID string) error

	// GetUserUrls возвращает список URL, сохранённых для указанного пользователя.
	GetUserUrls(uniqueID string) ([]storage.UserUrlsResponseBodyItem, error)

	// DeleteUserUrls удаляет указанный список коротких URL для указанного пользователя.
	DeleteUserUrls(uniqueID string, shortURLs []string) error

	// GetUsersCount получение количества пользователей
	GetUsersCount() (int, error)
}

// UserRepository реализует UserRepositoryInterface.
type UserRepository struct {
	Storage storage.UserStorageInterface
}

// IsUserExist проверяет, существует ли пользователь по уникальному ID.
func (ur *UserRepository) IsUserExist(uniqueID string) bool {
	return ur.Storage.IsUserExist(uniqueID)
}

// SaveUser сохраняет пользователя с указанным уникальным ID.
func (ur *UserRepository) SaveUser(uniqueID string) error {
	return ur.Storage.SaveUser(uniqueID)
}

// GetUserUrls возвращает список URL-адресов для указанного уникального ID пользователя.
func (ur *UserRepository) GetUserUrls(uniqueID string) ([]storage.UserUrlsResponseBodyItem, error) {
	return ur.Storage.GetUserUrls(uniqueID)
}

// DeleteUserUrls удаляет указанные URL-адреса для указанного уникального ID пользователя.
func (ur *UserRepository) DeleteUserUrls(uniqueID string, shortURLS []string) error {
	return ur.Storage.DeleteUserUrls(uniqueID, shortURLS)
}

// GetUsersCount получение количества пользователей
func (ur *UserRepository) GetUsersCount() (int, error) {
	return ur.Storage.GetUsersCount()
}
