package storage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"log"
)

const tableName = "urls"

type PgStorage struct {
	conn *pgx.Conn
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

func (pgs *PgStorage) GetShortURL(URL string) (string, bool) {
	query := fmt.Sprintf("SELECT short_url FROM %s WHERE url = $1", tableName)
	rows, err := pgs.conn.Query(pgs.ctx, query, URL)

	if err != nil {
		return "", false
	}

	var dataStorageRow DataStorageRow
	rowsCount := 0

	for rows.Next() {
		if err := rows.Scan(&dataStorageRow.ID, &dataStorageRow.ShortURL, &dataStorageRow.URL); err != nil {
			return "", false
		}

		rowsCount++
	}

	if rowsCount == 0 {
		return "", false
	}

	return dataStorageRow.ShortURL, true
}

func (pgs *PgStorage) Init(connectionString string) error {
	pgs.ctx = context.Background()
	var err error
	pgs.conn, err = pgx.Connect(pgs.ctx, connectionString)

	if err != nil {
		log.Fatalf("Error while initializing db connection: %v", err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		url VARCHAR(100),
		short_url VARCHAR(100)
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

func (pgs *PgStorage) Save(ShortURL string, URL string) error {
	query := fmt.Sprintf("INSERT INTO %s (short_url, url) VALUES ($1, $2)", tableName)
	_, err := pgs.conn.Exec(pgs.ctx, query, ShortURL, URL)
	return err
}

func (pgs *PgStorage) LoadData() ([]DataStorageRow, error) {
	return nil, nil
}

func (pgs *PgStorage) Close() {
	pgs.conn.Close(pgs.ctx)
}