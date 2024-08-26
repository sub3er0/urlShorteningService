package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
)

const tableName = "urls"

type PgStorage struct {
	conn *pgxpool.Pool
	ctx  context.Context
}

func (pgs *PgStorage) GetURL(shortURL string) (string, bool) {
	query := fmt.Sprintf("SELECT url FROM %s WHERE short_url = $1", tableName)
	rows, err := pgs.conn.Query(pgs.ctx, query, shortURL)

	if err != nil {
		return "", false
	}

	url := ""

	for rows.Next() {
		if err := rows.Scan(&url); err != nil {
			return url, false
		}
	}

	return url, true
}

func (pgs *PgStorage) GetURLCount() int {
	return 0
}

func (pgs *PgStorage) GetShortURL(URL string) (string, error) {
	query := fmt.Sprintf("SELECT short_url FROM %s WHERE url = $1", tableName)
	rows, err := pgs.conn.Query(pgs.ctx, query, URL)

	if err != nil {
		return "", err
	}

	shortURL := ""
	rowsCount := 0

	for rows.Next() {
		if err := rows.Scan(&shortURL); err != nil {
			return "", err
		}

		rowsCount++
	}

	if rowsCount == 0 {
		return "", errors.New("not found")
	}

	return shortURL, nil
}

func (pgs *PgStorage) Init(connectionString string) error {
	pgs.ctx = context.Background()
	var err error
	pgs.conn, err = pgxpool.Connect(pgs.ctx, connectionString)

	if err != nil {
		log.Fatalf("Error while initializing db connection: %v", err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		url VARCHAR(100) UNIQUE,
		short_url VARCHAR(100) UNIQUE,
	    user_id VARCHAR(100)
	);
	ALTER TABLE urls ADD UNIQUE (url, short_url);

	CREATE TABLE IF NOT EXISTS users_cookie (
		id SERIAL PRIMARY KEY,
		user_id VARCHAR(100) UNIQUE
	);`

	_, err = pgs.conn.Exec(pgs.ctx, createTableSQL)
	if err != nil {
		return err
	}

	return nil
}

func (pgs *PgStorage) Ping() bool {
	if err := pgs.conn.Ping(pgs.ctx); err != nil {
		return false
	}

	return true
}

func (pgs *PgStorage) Save(ShortURL string, URL string, userID string) error {
	query := fmt.Sprintf("INSERT INTO %s (short_url, url, user_id) VALUES ($1, $2, $3)", tableName)
	_, err := pgs.conn.Exec(pgs.ctx, query, ShortURL, URL, userID)
	return err
}

func (pgs *PgStorage) LoadData() ([]DataStorageRow, error) {
	return nil, nil
}

func (pgs *PgStorage) Close() {
	pgs.conn.Close()
}

func (pgs *PgStorage) SaveBatch(dataStorageRows []DataStorageRow) error {
	batch := &pgx.Batch{}
	for _, dataStorageRow := range dataStorageRows {
		batch.Queue(
			"INSERT INTO urls (url, short_url, user_id) VALUES ($1, $2, $3) ON CONFLICT (url, short_url) DO NOTHING",
			dataStorageRow.URL, dataStorageRow.ShortURL, dataStorageRow.UserID)
	}

	br := pgs.conn.SendBatch(context.Background(), batch)
	defer br.Close()

	for i := 0; i < len(dataStorageRows); i++ {
		_, err := br.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

func (pgs *PgStorage) IsUserExist(uniqueID string) bool {
	query := "SELECT id FROM users_cookie WHERE user_id = $1"
	rows, err := pgs.conn.Query(pgs.ctx, query, uniqueID)

	if err != nil {
		return false
	}

	var id int
	var rowsCount int

	for rows.Next() {
		if err := rows.Scan(&id); err != nil {
			return false
		}

		rowsCount++
	}

	return rowsCount > 0
}

func (pgs *PgStorage) SaveUser(uniqueID string) error {
	query := "INSERT INTO users_cookie (user_id) VALUES ($1)"
	_, err := pgs.conn.Exec(pgs.ctx, query, uniqueID)
	return err
}

func (pgs *PgStorage) GetUserUrls(uniqueID string) ([]UserUrlsResponseBodyItem, error) {
	query := fmt.Sprintf("SELECT url, short_url FROM %s WHERE user_id = $1", tableName)
	rows, err := pgs.conn.Query(pgs.ctx, query, uniqueID)

	if err != nil {
		return nil, err
	}

	var responseUrls []UserUrlsResponseBodyItem

	for rows.Next() {
		var url string
		var shortURL string
		var responseItem UserUrlsResponseBodyItem

		if err := rows.Scan(&url, &shortURL); err != nil {
			return nil, err
		}

		responseItem.OriginalURL = url
		responseItem.ShortURL = shortURL

		responseUrls = append(responseUrls, responseItem)
	}

	return responseUrls, nil
}
