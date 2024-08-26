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
	IsUserExist(uniqueId string) bool
	SaveUser(uniqueId string) error
	GetUserUrls(uniqueId string) ([]UserUrlsResponseBodyItem, error)
}
