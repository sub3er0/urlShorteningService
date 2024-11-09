package storage

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

// DataStorageRow представляет структуру для хранения информации о URL в хранилище.
// Эта структура используется для работы с сохранёнными данными пользователя в базе данных.
type DataStorageRow struct {
	ID          int    `json:"uuid"`         // Уникальный идентификатор записи
	ShortURL    string `json:"short_url"`    // Короткий URL
	URL         string `json:"original_url"` // Полный оригинальный URL
	UserID      string `json:"user_id"`      // Идентификатор пользователя, которому принадлежит запись
	DeletedFlag bool   `json:"is_deleted"`   // Флаг, указывающий, удалён ли URL
}

// UserUrlsResponseBodyItem представляет элемент ответа, содержащий информацию о URL пользователя.
// Эта структура используется при возвращении списка URL для пользователя.
type UserUrlsResponseBodyItem struct {
	OriginalURL string `json:"original_url"` // Полный оригинальный URL
	ShortURL    string `json:"short_url"`    // Короткий URL
}

// GetURLRow представляет результат, возвращаемый при получении длинного URL по короткому.
// Эта структура используется для обозначения состояния URL (например, удалён или активен).
type GetURLRow struct {
	URL       string // Полный URL
	IsDeleted bool   // Указывает, удалён ли URL
}

// DBConnectionInterface определяет методы для взаимодействия с базой данных.
// Этот интерфейс позволяет выполнять запросы, отправлять команды,
// проверять соединение и закрывать соединения.
type DBConnectionInterface interface {
	// Query выполняет SQL-запрос и возвращает строки результата.
	// Принятый контекст позволяет отменять запросы.
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)

	// Exec выполняет SQL-команду и возвращает тег команды,
	// указывающий на количество затронутых строк.
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)

	// Ping проверяет состояние подключения к базе данных.
	// Возвращает ошибку, если соединение недоступно.
	Ping(ctx context.Context) error

	// Close закрывает соединение с базой данных.
	Close()

	// SendBatch отправляет пакет запросов в базу данных.
	// Возвращает результаты отправленных батчей.
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}
