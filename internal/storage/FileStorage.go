package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
)

// FileStorage представляет хранилище данных в файловой системе.
// Она используется для сохранения и получения данных из файлов по заданному пути.
type FileStorage struct {
	// FileStoragePath указывает путь к файлу или директории, где будут храниться данные.
	FileStoragePath string
}

// SetConnection заглушка для интерфейса
func (fs *FileStorage) SetConnection(conn DBConnectionInterface) {}

// Init инициализирует хранилище данных. В этой реализации ничего не делает,
// так как данные хранятся в файловой системе.
func (fs *FileStorage) Init(connectionString string) error {
	return nil
}

// Close закрывает хранилище данных. В этом случае ничего не нужно делать,
// так как нет открытых ресурсов.
func (fs *FileStorage) Close() {}

// SaveBatch сохраняет пакет данных, представленных в виде массива DataStorageRow.
// Возвращает ошибку, если сохранение не удалось.
func (fs *FileStorage) SaveBatch(dataStorageRows []DataStorageRow) error {
	urlCount, err := fs.GetURLCount()
	if err != nil {
		return err
	}

	for i := range dataStorageRows {
		urlCount++
		dataStorageRows[i].ID = urlCount
	}

	var jsonRows []byte

	for _, dataStorageRow := range dataStorageRows {
		jsonRow, err := json.Marshal(dataStorageRow)
		jsonRows = append(jsonRows, jsonRow...)
		jsonRows = append(jsonRows, '\n')

		if err != nil {
			return err
		}
	}

	file, err := os.OpenFile(fs.FileStoragePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Printf("Error opening file:  %v\n", err)
		return err
	}

	defer file.Close()

	_, err = file.Write(jsonRows)
	if err != nil {
		return err
	}

	return nil
}

// GetURL возвращает полный URL для заданного короткого URL.
// Возвращает структуру GetURLRow и булевое значение, указывающее на существование.
func (fs *FileStorage) GetURL(shortURL string) (GetURLRow, bool) {
	file, err := os.OpenFile(fs.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
	var getURLRow GetURLRow

	if err != nil {
		return getURLRow, false
	}

	defer file.Close()
	reader := bufio.NewReader(file)
	var dataStorageRow DataStorageRow

	for {
		data, err := reader.ReadBytes('\n')

		if err == io.EOF {
			break
		}

		if err != nil {
			return getURLRow, false
		}

		err = json.Unmarshal(data, &dataStorageRow)

		if err != nil {
			return getURLRow, false
		}

		getURLRow.URL = dataStorageRow.URL
		if dataStorageRow.ShortURL == shortURL {
			return getURLRow, true
		}
	}

	return getURLRow, false
}

// GetURLCount возвращает количество сохранённых URL в хранилище.
func (fs *FileStorage) GetURLCount() (int, error) {
	file, err := os.OpenFile(fs.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)

	if err != nil {
		return 0, err
	}

	defer file.Close()
	reader := bufio.NewReader(file)
	count := 0

	for {
		_, err := reader.ReadBytes('\n')

		if err == io.EOF {
			break
		}

		count++
	}

	return count, nil
}

// GetShortURL ищет короткий URL для заданного оригинального URL.
// Возвращает короткий URL, если он найден, и ошибку, если нет.
func (fs *FileStorage) GetShortURL(URL string) (string, error) {
	file, err := os.OpenFile(fs.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)

	if err != nil {
		return "", err
	}

	defer file.Close()
	reader := bufio.NewReader(file)
	var dataStorageRow DataStorageRow

	for {
		data, readError := reader.ReadBytes('\n')

		if readError == io.EOF {
			break
		}

		if readError != nil {
			return "", readError
		}

		readError = json.Unmarshal(data, &dataStorageRow)

		if readError != nil {
			return "", readError
		}

		if dataStorageRow.URL == URL {
			return dataStorageRow.ShortURL, nil
		}
	}

	err = errors.New("short url not found")
	return "", err
}

// Save сохраняет короткий URL с соответствующим полному URL и идентификатору пользователя.
// Параметры:
//   - ShortURL: короткий URL, который должен быть сохранён.
//   - URL: полный (оригинальный) URL, который связан с коротким URL.
//   - userID: идентификатор пользователя, который добавляет URL.
//
// Возвращает ошибку, если сохранение не удалось.
func (fs *FileStorage) Save(ShortURL string, URL string, userID string) error {
	ID, err := fs.GetURLCount()

	if err != nil {
		return err
	}

	row := DataStorageRow{
		ID:       ID,
		ShortURL: ShortURL,
		URL:      URL,
	}
	jsonRow, err := json.Marshal(row)

	if err != nil {
		return err
	}

	file, err := os.OpenFile(fs.FileStoragePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Printf("Error opening file:  %v\n", err)
		return err
	}

	defer file.Close()
	jsonRow = append(jsonRow, '\n')
	_, err = file.Write(jsonRow)
	if err != nil {
		return err
	}

	return nil
}

// LoadData загружает данные из хранилища и возвращает их в виде массива DataStorageRow.
// Возвращает массив DataStorageRow и ошибку, если произошла ошибка чтения данных.
func (fs *FileStorage) LoadData() ([]DataStorageRow, error) {
	file, err := os.OpenFile(fs.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)

	if err != nil {
		return nil, err
	}

	defer file.Close()
	reader := bufio.NewReader(file)
	var dataStorageRow DataStorageRow
	var dataStorageRows []DataStorageRow

	for {
		data, err := reader.ReadBytes('\n')

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(data, &dataStorageRow)

		if err != nil {
			return nil, err
		}

		dataStorageRows = append(dataStorageRows, dataStorageRow)
	}

	return dataStorageRows, nil
}

// Ping проверяет состояние работы хранилища.
// Возвращает true, так как хранилище работает в оперативной памяти.
func (fs *FileStorage) Ping() bool {
	return true
}

// IsUserExist проверяет, существует ли пользователь по уникальному идентификатору.
// В данной реализации всегда возвращает false, так как InMemoryStorage не хранит пользователей.
func (fs *FileStorage) IsUserExist(data string) bool {
	return false
}

// SaveUser сохраняет нового пользователя с указанным уникальным идентификатором.
// В данной реализации ничего не делает и всегда возвращает nil.
func (fs *FileStorage) SaveUser(uniqueID string) error {
	return nil
}

// GetUserUrls возвращает список URL, сохранённых для указанного пользователя.
// В данной реализации всегда возвращает nil, так как InMemoryStorage не хранит пользователей.
func (fs *FileStorage) GetUserUrls(uniqueID string) ([]UserUrlsResponseBodyItem, error) {
	return nil, nil
}

// DeleteUserUrls удаляет указанные короткие URL для данного пользователя.
// В данной реализации ничего не делает и всегда возвращает nil.
func (fs *FileStorage) DeleteUserUrls(uniqueID string, shortURLS []string) error {
	return nil
}

// GetUsersCount получение количества пользователей
func (fs *FileStorage) GetUsersCount() (int, error) {
	return 0, nil
}
