package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// UserStorageInterface определяет методы для работы с хранилищем пользователей.
// Этот интерфейс предоставляет доступ к операциям проверки существования пользователя,
// сохранения пользователей, получения их URL и удаления URL.
type UserStorageInterface interface {
	// IsUserExist проверяет, существует ли пользователь по указанному уникальному идентификатору.
	// Возвращает true, если пользователь существует, и false в противном случае.
	IsUserExist(uniqueID string) bool

	// SaveUser сохраняет нового пользователя с заданным уникальным идентификатором.
	// Возвращает ошибку, если сохранение не удалось.
	SaveUser(uniqueID string) error

	// GetUserUrls возвращает список коротких URL для указанного пользователя.
	// В случае успешного получения возвращает массив UserUrlsResponseBodyItem и nil.
	// В случае ошибки возвращает nil и ошибку.
	GetUserUrls(uniqueID string) ([]UserUrlsResponseBodyItem, error)

	// DeleteUserUrls удаляет указанные короткие URL для указанного пользователя.
	// Возвращает ошибку, если возникла ошибка удаления.
	DeleteUserUrls(uniqueID string, shortURLs []string) error

	// Init инициализирует хранилище пользователей с помощью строки соединения.
	// Возвращает ошибку, если произошла ошибка инициализации.
	Init(connectionString string) error

	// Close закрывает соединение с хранилищем данных.
	Close()
}

// UsersStorage предоставляет реализацию для работы с хранилищем пользователей
// и взаимодействия с базой данных через пул соединений pgx.
type UsersStorage struct {
	// conn представляет пул соединений с базой данных, позволяющий выполнять SQL-команды и запросы.
	conn *pgxpool.Pool

	// ctx представляет контекст, используемый для управления временем жизни запросов и операций.
	ctx context.Context
}

// IsUserExist проверяет, существует ли пользователь по его уникальному идентификатору.
// Возвращает true, если пользователь существует, и false в противном случае.
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

// SaveUser сохраняет нового пользователя с указанным уникальным идентификатором.
// Возвращает ошибку, если сохранение не удалось.
func (us *UsersStorage) SaveUser(uniqueID string) error {
	query := "INSERT INTO users_cookie (user_id) VALUES ($1)"
	_, err := us.conn.Exec(us.ctx, query, uniqueID)
	return err
}

// GetUserUrls возвращает список URL, сохраненных для указанного пользователя.
// Возвращает массив UserUrlsResponseBodyItem и ошибку, если произошла ошибка чтения.
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

// DeleteUserUrls удаляет указанные короткие URL для указанного пользователя.
// Возвращает ошибку, если возникла ошибка удаления.
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

// Init инициализирует соединение с базой данных по заданной строке подключения.
// Параметры:
//   - connectionString: строка подключения к базе данных.
//
// Возвращает ошибку, если инициализация соединения не удалась.
func (us *UsersStorage) Init(connectionString string) error {
	us.ctx = context.Background()
	var err error
	us.conn, err = pgxpool.Connect(us.ctx, connectionString)

	if err != nil {
		log.Fatalf("Error while initializing db connection: %v", err)
	}

	return nil
}

// Close закрывает соединение с базой данных.
// Этот метод должен вызываться для освобождения всех ресурсов, занимаемых соединением.
func (us *UsersStorage) Close() {
	us.conn.Close()
}
