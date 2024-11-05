package storage

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

// DefaultStorage предоставляет реализацию для работы с хранилищем данных.
// Она включает в себя соединение с базой данных и контекст для управления операциям.
type DefaultStorage struct {
	// conn представляет подключение к базе данных, позволяющее выполнять команды и запросы.
	conn DBConnectionInterface

	// ctx представляет контекст, который используется для управления временем жизни запросов и операций.
	ctx context.Context
}

// Init инициализирует соединение с базой данных по заданной строке подключения.
// Параметры:
//   - connectionString: строка подключения к базе данных.
//
// Возвращает ошибку, если инициализация соединения не удалась или
// если возникла ошибка при выполнении SQL-команд.
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

// Close закрывает соединение с базой данных.
// Этот метод должен вызываться для освобождения всех ресурсов, занимаемых соединением.
func (ds *DefaultStorage) Close() {
	ds.conn.Close()
}
