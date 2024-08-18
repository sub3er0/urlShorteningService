package storage

// URLStorage Интерфейс для хранения URL-адресов
type URLStorage interface {
	GetURL(shortURL string) (string, bool)
	GetURLCount() int
	GetShortURL(URL string) (string, bool)
	Save(ShortURL string, URL string) error
	LoadData() ([]DataStorageRow, error)
	Ping() bool
	SaveBatch(dataStorageRows []DataStorageRow) error
}
