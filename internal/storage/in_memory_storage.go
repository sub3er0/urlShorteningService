package storage

import (
	"errors"
)

// InMemoryStorage Пример реализации хранения в памяти
type InMemoryStorage struct {
	Urls map[string]string
}

func (ims *InMemoryStorage) SaveBatch(dataStorageRows []DataStorageRow) error {
	for _, row := range dataStorageRows {
		ims.Urls[row.ShortURL] = row.URL
	}
	return nil
}

func (ims *InMemoryStorage) Save(ShorURL string, URL string, userID string) error {
	ims.Urls[ShorURL] = URL
	return nil
}

func (ims *InMemoryStorage) LoadData() ([]DataStorageRow, error) {
	return make([]DataStorageRow, 0), nil
}

func (ims *InMemoryStorage) GetURL(shortURL string) (GetURLRow, bool) {
	var getUrlRow GetURLRow
	var ok bool
	getUrlRow.URL, ok = ims.Urls[shortURL]

	return getUrlRow, ok
}

func (ims *InMemoryStorage) GetURLCount() int {
	return len(ims.Urls)
}

func (ims *InMemoryStorage) GetShortURL(URL string) (string, error) {
	err := errors.New("short url not found")

	for k, v := range ims.Urls {
		if v == URL {
			return k, nil
		}
	}

	return "", err
}

func (ims *InMemoryStorage) Set(shortURL, longURL string) error {
	ims.Urls[shortURL] = longURL
	return nil
}

func (ims *InMemoryStorage) Ping() bool {
	return true
}

func (ims *InMemoryStorage) IsUserExist(data string) bool {
	return false
}

func (ims *InMemoryStorage) SaveUser(uniqueID string) error {
	return nil
}

func (ims *InMemoryStorage) GetUserUrls(uniqueID string) ([]UserUrlsResponseBodyItem, error) {
	return nil, nil
}

func (ims *InMemoryStorage) DeleteUserUrls(uniqueID string, shortURLS []string) error {
	return nil
}
