package storage

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"os"
)

type DataStorageRow struct {
	ID       int    `json:"uuid"`
	ShortURL string `json:"short_url"`
	URL      string `json:"original_url"`
}

type FileStorage struct {
	FileStoragePath string
}

func (fs *FileStorage) GetURL(shortURL string) (string, bool) {
	file, err := os.OpenFile(fs.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)

	if err != nil {
		return "", false
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
			return "", false
		}

		err = json.Unmarshal(data, &dataStorageRow)

		if err != nil {
			return "", false
		}

		if dataStorageRow.ShortURL == shortURL {
			return dataStorageRow.URL, true
		}
	}

	return "", false
}

func (fs *FileStorage) GetURLCount() int {
	file, err := os.OpenFile(fs.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)

	if err != nil {
		return 0
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

	return count
}

func (fs *FileStorage) GetShortURL(URL string) (string, bool) {
	file, err := os.OpenFile(fs.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)

	if err != nil {
		return "", false
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
			return "", false
		}

		err = json.Unmarshal(data, &dataStorageRow)

		if err != nil {
			return "", false
		}

		if dataStorageRow.URL == URL {
			return dataStorageRow.ShortURL, true
		}
	}

	return "", false
}

func (fs *FileStorage) Save(ShortURL string, URL string) error {
	row := DataStorageRow{
		ID:       fs.GetURLCount(),
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

func (fs *FileStorage) Ping() bool {
	return true
}
