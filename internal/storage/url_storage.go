package storage

// URLStorage Интерфейс для хранения URL-адресов
type URLStorage interface {
	GetURL(shortURL string) (string, bool)
	GetUrlCount() int
	GetShortURL(URL string) (string, bool)
	Set(shortURL, longURL string) error
}
