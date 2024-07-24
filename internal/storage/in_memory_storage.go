package storage

// InMemoryStorage Пример реализации хранения в памяти
type InMemoryStorage struct {
	Urls map[string]string
}

func (ims *InMemoryStorage) GetURL(shortURL string) (string, bool) {
	longURL, ok := ims.Urls[shortURL]
	return longURL, ok
}

func (ims *InMemoryStorage) GetURLCount() int {
	return len(ims.Urls)
}

func (ims *InMemoryStorage) GetShortURL(URL string) (string, bool) {
	for k, v := range ims.Urls {
		if v == URL {
			return k, true
		}
	}

	return "", false
}

func (ims *InMemoryStorage) Set(shortURL, longURL string) error {
	ims.Urls[shortURL] = longURL
	return nil
}
