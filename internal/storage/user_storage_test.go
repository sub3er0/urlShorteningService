package storage

import (
	"context"
	"errors"
	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUsersStorage_IsUserExist(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	ctx := context.Background()
	storage := &UsersStorage{conn: mock, ctx: ctx}

	uniqueID := "user123"

	// Установка ожидания для SQL запроса
	mock.ExpectQuery("SELECT id FROM users_cookie WHERE user_id = \\$1").
		WithArgs(uniqueID).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1)) // Успешный случай, пользователь существует

	exists := storage.IsUserExist(uniqueID)
	assert.True(t, exists, "Expected user to exist")
	assert.NoError(t, mock.ExpectationsWereMet(), "There should be no unfulfilled expectations")

	// Проверяем случай, когда пользователь не существует
	mock.ExpectQuery("SELECT id FROM users_cookie WHERE user_id = \\$1").
		WithArgs(uniqueID).
		WillReturnRows(pgxmock.NewRows([]string{"id"})) // Пользователь не найден

	exists = storage.IsUserExist(uniqueID)
	assert.False(t, exists, "Expected user to not exist")
	assert.NoError(t, mock.ExpectationsWereMet(), "There should be no unfulfilled expectations")
}

func TestUsersStorage_SaveUser(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	ctx := context.Background()
	storage := &UsersStorage{conn: mock, ctx: ctx}

	uniqueID := "user123"

	// Установка ожидания для SQL запроса
	mock.ExpectExec("INSERT INTO users_cookie \\(user_id\\) VALUES \\(\\$1\\)").
		WithArgs(uniqueID).
		WillReturnResult(pgxmock.NewResult("1", 1)) // Успешная вставка

	err = storage.SaveUser(uniqueID)
	assert.NoError(t, err, "Expected no error during SaveUser")
	assert.NoError(t, mock.ExpectationsWereMet(), "There should be no unfulfilled expectations")
}

func TestUsersStorage_GetUserUrls(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	ctx := context.Background()
	storage := &UsersStorage{conn: mock, ctx: ctx}

	uniqueID := "user123"

	// Установка ожидания для SQL запроса
	mock.ExpectQuery("SELECT url, short_url FROM urls WHERE user_id = \\$1 AND is_deleted = false").
		WithArgs(uniqueID).
		WillReturnRows(pgxmock.NewRows([]string{"url", "short_url"}).
			AddRow("http://example.com", "short.ly/xyz").
			AddRow("http://example2.com", "short.ly/abc")) // Данные для пользователя

	urls, err := storage.GetUserUrls(uniqueID)
	assert.NoError(t, err, "Expected no error during GetUserUrls")
	assert.Len(t, urls, 2, "Expected 2 URLs to be returned")
	assert.Equal(t, "http://example.com", urls[0].OriginalURL, "Expected first URL to match")
	assert.Equal(t, "short.ly/xyz", urls[0].ShortURL, "Expected first short URL to match")
	assert.NoError(t, mock.ExpectationsWereMet(), "There should be no unfulfilled expectations")

	// Проверяем случай, когда возникает ошибка
	mock.ExpectQuery("SELECT url, short_url FROM urls WHERE user_id = \\$1 AND is_deleted = false").
		WithArgs(uniqueID).
		WillReturnError(errors.New("query error")) // Ошибка выполнения запроса

	urls, err = storage.GetUserUrls(uniqueID)
	assert.Error(t, err, "Expected error during GetUserUrls")
	assert.Nil(t, urls, "Expected nil URLs in case of error")
	assert.NoError(t, mock.ExpectationsWereMet(), "There should be no unfulfilled expectations")
}
