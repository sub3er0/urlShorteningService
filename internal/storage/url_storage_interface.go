package storage

// URLStorage Интерфейс для хранения URL-адресов
type URLStorage interface {
	GetURL(shortURL string) (string, bool)
	GetURLCount() int
	GetShortURL(URL string) (string, error)
	Save(ShortURL string, URL string, userID string) error
	LoadData() ([]DataStorageRow, error)
	Ping() bool
	SaveBatch(dataStorageRows []DataStorageRow) error
	IsUserExist(uniqueID string) bool
	SaveUser(uniqueID string) error
	GetUserUrls(uniqueID string) ([]UserUrlsResponseBodyItem, error)
}
