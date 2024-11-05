package repository

import "github.com/sub3er0/urlShorteningService/internal/storage"

// URLRepositoryInterface определяет методы для работы с репозиторием URL.
// Этот интерфейс предоставляет доступ к операциям получения, сохранения и манипуляции с URL в хранилище.
type URLRepositoryInterface interface {
	// GetURL возвращает полный URL и статус его наличия по заданному короткому URL.
	GetURL(shortURL string) (storage.GetURLRow, bool)

	// GetURLCount возвращает общее количество URL в репозитории.
	GetURLCount() int

	// GetShortURL возвращает короткий URL для заданного полного URL.
	// Если в репозитории нет запись, возвращается ошибка.
	GetShortURL(URL string) (string, error)

	// Save сохраняет короткий URL с соответствующим полному URL и идентификатором пользователя.
	Save(ShortURL string, URL string, userID string) error

	// LoadData загружает данные о URL из хранилища в виде массива DataStorageRow.
	LoadData() ([]storage.DataStorageRow, error)

	// Ping проверяет состояние соединения с базой данных.
	// Возвращает true, если соединение успешно.
	Ping() bool

	// SaveBatch сохраняет пакет данных, представленных в виде массива DataStorageRow.
	SaveBatch(dataStorageRows []storage.DataStorageRow) error
}

// URLRepository отвечает за взаимодействие между
// бизнес-логикой приложения и хранилищем данных URL.
// Он инкапсулирует методы для работы с хранения и получения URL.
type URLRepository struct {
	// Storage представляет собой интерфейс для взаимодействия с хранилищем URL.
	Storage storage.URLStorageInterface
}

// GetStorage возвращает текущее хранилище URL, используемое в репозитории.
func (ur *URLRepository) GetStorage() storage.URLStorageInterface {
	return ur.Storage
}

// GetURL получает URL по его короткому формату.
func (ur *URLRepository) GetURL(shortURL string) (storage.GetURLRow, bool) {
	return ur.Storage.GetURL(shortURL)
}

// GetURLCount возвращает количество URL.
func (ur *URLRepository) GetURLCount() int {
	return ur.Storage.GetURLCount()
}

// GetShortURL возвращает короткий URL, если он существует.
func (ur *URLRepository) GetShortURL(URL string) (string, error) {
	return ur.Storage.GetShortURL(URL)
}

// Save сохраняет короткий URL и оригинальный URL для пользователя.
func (ur *URLRepository) Save(ShortURL string, URL string, userID string) error {
	return ur.Storage.Save(ShortURL, URL, userID)
}

// LoadData загружает данные из хранилища.
func (ur *URLRepository) LoadData() ([]storage.DataStorageRow, error) {
	return ur.Storage.LoadData()
}

// Ping проверяет доступность хранилища.
func (ur *URLRepository) Ping() bool {
	return ur.Storage.Ping()
}

// SaveBatch сохраняет пакет данных в хранилище.
func (ur *URLRepository) SaveBatch(dataStorageRows []storage.DataStorageRow) error {
	return ur.Storage.SaveBatch(dataStorageRows)
}
