package repository

import "github.com/sub3er0/urlShorteningService/internal/storage"

type UrlRepositoryInterface interface {
	GetURL(shortURL string) (storage.GetURLRow, bool)
	GetURLCount() int
	GetShortURL(URL string) (string, error)
	Save(ShortURL string, URL string, userID string) error
	LoadData() ([]storage.DataStorageRow, error)
	Ping() bool
	SaveBatch(dataStorageRows []storage.DataStorageRow) error
}

type UrlRepository struct {
	Storage storage.URLStorageInterface
}

func (ur *UrlRepository) GetStorage() storage.URLStorageInterface {
	return ur.Storage
}

func (ur *UrlRepository) GetURL(shortURL string) (storage.GetURLRow, bool) {
	return ur.Storage.GetURL(shortURL)
}

// GetURLCount возвращает количество URL.
func (ur *UrlRepository) GetURLCount() int {
	return ur.Storage.GetURLCount()
}

// GetShortURL возвращает короткий URL, если он существует.
func (ur *UrlRepository) GetShortURL(URL string) (string, error) {
	return ur.Storage.GetShortURL(URL)
}

// Save сохраняет короткий URL и оригинальный URL для пользователя.
func (ur *UrlRepository) Save(ShortURL string, URL string, userID string) error {
	return ur.Storage.Save(ShortURL, URL, userID)
}

// LoadData загружает данные из хранилища.
func (ur *UrlRepository) LoadData() ([]storage.DataStorageRow, error) {
	return ur.Storage.LoadData()
}

// Ping проверяет доступность хранилища.
func (ur *UrlRepository) Ping() bool {
	return ur.Storage.Ping()
}

// SaveBatch сохраняет пакет данных в хранилище.
func (ur *UrlRepository) SaveBatch(dataStorageRows []storage.DataStorageRow) error {
	return ur.Storage.SaveBatch(dataStorageRows)
}
