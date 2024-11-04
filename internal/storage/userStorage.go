package storage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
)

type UserStorageInterface interface {
	IsUserExist(uniqueID string) bool
	SaveUser(uniqueID string) error
	GetUserUrls(uniqueID string) ([]UserUrlsResponseBodyItem, error)
	DeleteUserUrls(uniqueID string, shortURLS []string) error
	Init(connectionString string) error
	Close()
}

type UsersStorage struct {
	conn *pgxpool.Pool
	ctx  context.Context
}

func (us *UsersStorage) IsUserExist(uniqueID string) bool {
	query := "SELECT id FROM users_cookie WHERE user_id = $1"
	rows, err := us.conn.Query(us.ctx, query, uniqueID)

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

func (us *UsersStorage) SaveUser(uniqueID string) error {
	query := "INSERT INTO users_cookie (user_id) VALUES ($1)"
	_, err := us.conn.Exec(us.ctx, query, uniqueID)
	return err
}

func (us *UsersStorage) GetUserUrls(uniqueID string) ([]UserUrlsResponseBodyItem, error) {
	query := fmt.Sprintf("SELECT url, short_url FROM %s WHERE user_id = $1 AND is_deleted = false", tableName)
	rows, err := us.conn.Query(us.ctx, query, uniqueID)

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

func (us *UsersStorage) DeleteUserUrls(uniqueID string, shortURLS []string) error {
	batch := &pgx.Batch{}
	for _, shortURL := range shortURLS {
		batch.Queue(
			"UPDATE urls SET is_deleted = true WHERE short_url = $1 AND user_id = $2", shortURL, uniqueID)
	}

	br := us.conn.SendBatch(context.Background(), batch)
	defer br.Close()

	for i := 0; i < len(shortURLS); i++ {
		_, err := br.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

func (us *UsersStorage) Init(connectionString string) error {
	us.ctx = context.Background()
	var err error
	us.conn, err = pgxpool.Connect(us.ctx, connectionString)

	if err != nil {
		log.Fatalf("Error while initializing db connection: %v", err)
	}

	return nil
}

func (us *UsersStorage) Close() {
	us.conn.Close()
}
