package storage

// URLStorage Интерфейс для хранения URL-адресов
type URLStorage interface {
	GetURL(shortURL string) (string, bool)
	GetShortURL(Url string) (string, bool)
	Set(shortURL, longURL string) error
}
