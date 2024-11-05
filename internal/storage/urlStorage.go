package storage

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// URLStorageInterface определяет методы для работы с хранилищем URL.
type URLStorageInterface interface {
	// GetURL получит URL по его короткому формату.
	GetURL(shortURL string) (GetURLRow, bool)

	// GetURLCount возвращает общее количество URL в хранилище.
	GetURLCount() int

	// GetShortURL возвращает короткий формат URL для заданного полного URL.
	GetShortURL(URL string) (string, error)

	// Save сохраняет короткий URL с соответствующим полным URL и идентификатором пользователя.
	Save(ShortURL string, URL string, userID string) error

	// LoadData загружает данные из хранилища в массив DataStorageRow.
	LoadData() ([]DataStorageRow, error)

	// Ping проверяет состояние подключения к базе данных.
	Ping() bool

	// SaveBatch сохраняет пакет данных, представленных в виде массива DataStorageRow.
	SaveBatch(dataStorageRows []DataStorageRow) error

	// Init инициализирует соединение с хранилищем данных, используя заданную строку подключения.
	Init(connectionString string) error

	// Close закрывает соединение с хранилищем данных.
	Close()
}

// CommandTag - интерфейс для работы с результатами выполнения команд.
type CommandTag interface{}

// URLStorage предоставляет методы для работы с хранилищем URL в базе данных.
// Она управляет соединением с базой данных и контекстом, необходимым для выполнения операций.
type URLStorage struct {
	// conn представляет соединение с базой данных, предоставляющее доступ к методам SQL.
	conn DBConnectionInterface

	// ctx представляет контекст, используемый для управления временем жизни запросов и операций.
	ctx context.Context
}

// GetURL возвращает строку, соответствующую заданному короткому URL.
// Возвращает структуру GetURLRow и булевое значение, указывающее на успех или неудачу.
func (us *URLStorage) GetURL(shortURL string) (GetURLRow, bool) {
	var getURLRow GetURLRow
	query := fmt.Sprintf("SELECT url, is_deleted FROM %s WHERE short_url = $1", tableName)
	rows, err := us.conn.Query(us.ctx, query, shortURL)

	if err != nil {
		return getURLRow, false
	}

	rowsCount := 0

	for rows.Next() {
		if err := rows.Scan(&getURLRow.URL, &getURLRow.IsDeleted); err != nil {
			return getURLRow, false
		}

		rowsCount++
	}

	if rowsCount == 0 {
		return getURLRow, false
	}

	return getURLRow, true
}

// GetURLCount возвращает общее количество URL в хранилище.
func (us *URLStorage) GetURLCount() int {
	return 0
}

// GetShortURL возвращает короткий URL для указанного полного URL.
// Если в репозитории не найдено, возвращает ошибку.
func (us *URLStorage) GetShortURL(URL string) (string, error) {
	query := fmt.Sprintf("SELECT short_url FROM %s WHERE url = $1", tableName)
	rows, err := us.conn.Query(us.ctx, query, URL)

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

// Init инициализирует соединение с базой данных по заданной строке подключения.
func (us *URLStorage) Init(connectionString string) error {
	us.ctx = context.Background()
	var err error
	us.conn, err = pgxpool.Connect(us.ctx, connectionString)

	if err != nil {
		log.Fatalf("Error while initializing db connection: %v", err)
	}

	return nil
}

// Ping проверяет состояние соединения с базой данных.
// Возвращает true, если соединение успешно, и false, если возникает ошибка.
func (us *URLStorage) Ping() bool {
	if err := us.conn.Ping(us.ctx); err != nil {
		return false
	}

	return true
}

// Save сохраняет короткий URL с соответствующим полному URL и идентификатором пользователя.
func (us *URLStorage) Save(ShortURL string, URL string, userID string) error {
	query := fmt.Sprintf("INSERT INTO %s (short_url, url, user_id) VALUES ($1, $2, $3)", tableName)
	_, err := us.conn.Exec(us.ctx, query, ShortURL, URL, userID)
	return err
}

// LoadData загружает данные из хранилища и возвращает их.
func (us *URLStorage) LoadData() ([]DataStorageRow, error) {
	return nil, nil
}

// Close закрывает соединение с базой данных.
func (us *URLStorage) Close() {
	us.conn.Close()
}

// SaveBatch сохраняет пакетные данные, представленные в виде массива DataStorageRow.
func (us *URLStorage) SaveBatch(dataStorageRows []DataStorageRow) error {
	batch := &pgx.Batch{}
	for _, dataStorageRow := range dataStorageRows {
		batch.Queue(
			"INSERT INTO urls (url, short_url, user_id) VALUES ($1, $2, $3) ON CONFLICT (url, short_url) DO NOTHING",
			dataStorageRow.URL, dataStorageRow.ShortURL, dataStorageRow.UserID)
	}

	br := us.conn.SendBatch(context.Background(), batch)
	defer br.Close()

	for i := 0; i < len(dataStorageRows); i++ {
		_, err := br.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}
