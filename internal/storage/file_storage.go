package storage

import (
	"encoding/json"
	"log"
	"os"
)

type FileStorageRow struct {
	Id       int    `json:"uuid"`
	ShortURL string `json:"short_url"`
	URL      string `json:"original_url"`
}

func Save(row FileStorageRow, fileStoragePath string) {
	jsonRow, err := json.Marshal(row)

	if err != nil {
		log.Fatalf("Serialization fail: %v", err)
	}

	file, err := os.OpenFile(fileStoragePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Fatalf("Error opening file:  %v", err)
	}

	defer file.Close()
	jsonRow = append(jsonRow, '\n')
	_, err = file.Write(jsonRow)
	if err != nil {
		log.Fatalf("Error writing to file:  %v", err)
	}
}
