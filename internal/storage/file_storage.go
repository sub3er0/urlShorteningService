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

func (fs *FileStorage) Save(row DataStorageRow) error {
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
