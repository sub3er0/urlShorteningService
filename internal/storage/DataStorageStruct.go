package storage

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type DataStorageRow struct {
	ID          int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	URL         string `json:"original_url"`
	UserID      string `json:"user_id"`
	DeletedFlag bool   `json:"is_deleted"`
}

type UserUrlsResponseBodyItem struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
}

type GetURLRow struct {
	URL       string
	IsDeleted bool
}

type DBConnectionInterface interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Ping(ctx context.Context) error
	Close()
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

//// Rows - интерфейс для обработки возвращаемых строк запроса.
//type Rows interface {
//	Next() bool                     // Метод для перехода к следующей строке
//	Scan(dest ...interface{}) error // Метод для сканирования текущей строки
//	Close() error                   // Метод закрытия результатов
//}
