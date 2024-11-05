package storage

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
)

type DefaultStorage struct {
	conn DBConnectionInterface
	ctx  context.Context
}

func (ds *DefaultStorage) Init(connectionString string) error {
	ds.ctx = context.Background()
	var err error
	ds.conn, err = pgxpool.Connect(ds.ctx, connectionString)

	if err != nil {
		log.Fatalf("Error while initializing db connection: %v", err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		url VARCHAR(100) UNIQUE,
		short_url VARCHAR(100) UNIQUE,
	    user_id VARCHAR(100),
	    is_deleted BOOLEAN DEFAULT FALSE
	);
	ALTER TABLE urls ADD UNIQUE (url, short_url);

	CREATE TABLE IF NOT EXISTS users_cookie (
		id SERIAL PRIMARY KEY,
		user_id VARCHAR(100) UNIQUE
	);`

	_, err = ds.conn.Exec(ds.ctx, createTableSQL)
	if err != nil {
		return err
	}

	return nil
}

func (ds *DefaultStorage) Close() {
	ds.conn.Close()
}
