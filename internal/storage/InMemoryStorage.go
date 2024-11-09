package storage

import (
	"errors"
)

// InMemoryStorage Пример реализации хранения в памяти
type InMemoryStorage struct {
	Urls map[string]string
}

// Init инициализирует хранилище. В данной реализации ничего не делает,
// так как хранилище работает в оперативной памяти.
func (ims *InMemoryStorage) Init(connectionString string) error {
	return nil
}

// Close закрывает хранилище. В данной реализации ничего не делает.
func (ims *InMemoryStorage) Close() {}

// SaveBatch сохраняет пакет данных, представленных в виде массива DataStorageRow.
// Возвращает ошибку, если не удалось сохранить данные.
func (ims *InMemoryStorage) SaveBatch(dataStorageRows []DataStorageRow) error {
	for _, row := range dataStorageRows {
		ims.Urls[row.ShortURL] = row.URL
	}
	return nil
}

// Save сохраняет новый короткий URL с соответствующим полному URL и идентификатору пользователя.
// Возвращает ошибку, если не удалось сохранить данные.
func (ims *InMemoryStorage) Save(ShorURL string, URL string, userID string) error {
	ims.Urls[ShorURL] = URL
	return nil
}

// LoadData загружает данные из хранилища и возвращает их в виде массива DataStorageRow.
// В данной реализации просто возвращает пустой массив.
func (ims *InMemoryStorage) LoadData() ([]DataStorageRow, error) {
	return make([]DataStorageRow, 0), nil
}

// GetURL возвращает URL для заданного короткого URL.
// Возвращает структуру GetURLRow и булевое значение, указывающее на существование.
func (ims *InMemoryStorage) GetURL(shortURL string) (GetURLRow, bool) {
	var getURLRow GetURLRow
	var ok bool
	getURLRow.URL, ok = ims.Urls[shortURL]

	return getURLRow, ok
}

// GetURLCount возвращает количество сохранённых URL в хранилище.
func (ims *InMemoryStorage) GetURLCount() int {
	return len(ims.Urls)
}

// GetShortURL ищет короткий URL для заданного оригинального URL.
// Возвращает короткий URL, если он найден, и ошибку, если нет.
func (ims *InMemoryStorage) GetShortURL(URL string) (string, error) {
	err := errors.New("short url not found")

	for k, v := range ims.Urls {
		if v == URL {
			return k, nil
		}
	}

	return "", err
}

// Set добавляет данные в хранилище
func (ims *InMemoryStorage) Set(shortURL, longURL string) error {
	ims.Urls[shortURL] = longURL
	return nil
}

// Ping проверяет состояние работы хранилища.
// Возвращает true, так как хранилище работает в оперативной памяти.
func (ims *InMemoryStorage) Ping() bool {
	return true
}

// IsUserExist проверяет, существует ли пользователь по уникальному идентификатору.
// В данной реализации всегда возвращает false.
func (ims *InMemoryStorage) IsUserExist(data string) bool {
	return false
}

// SaveUser сохраняет нового пользователя с указанным уникальным идентификатором.
// В данной реализации ничего не делает и всегда возвращает nil.
func (ims *InMemoryStorage) SaveUser(uniqueID string) error {
	return nil
}

// GetUserUrls возвращает список URL, сохранённых для указанного пользователя.
// В данной реализации всегда возвращает nil.
func (ims *InMemoryStorage) GetUserUrls(uniqueID string) ([]UserUrlsResponseBodyItem, error) {
	return nil, nil
}

// DeleteUserUrls удаляет указанные короткие URL для данного пользователя.
// В данной реализации ничего не делает и всегда возвращает nil.
func (ims *InMemoryStorage) DeleteUserUrls(uniqueID string, shortURLS []string) error {
	return nil
}
