package storage

import (
	"context"
	"testing"

	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
)

func TestURLStorage_GetURL(t *testing.T) {
	// Создаем мок объекта pgxpool
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	ctx := context.Background()
	storage := &URLStorage{conn: mock, ctx: ctx}

	// Подготовка тестовых данных
	shortURL := "short.ly/xyz"
	expectedURL := "http://example.com"
	expectedIsDeleted := false

	// Задаем ожидание для SQL запроса
	mock.ExpectQuery(`SELECT url, is_deleted FROM urls WHERE short_url = \$1`).
		WithArgs(shortURL).
		WillReturnRows(pgxmock.NewRows([]string{"url", "is_deleted"}).
			AddRow(expectedURL, expectedIsDeleted))

	// Выполнение теста
	urlRow, ok := storage.GetURL(shortURL)

	// Проверка результатов
	assert.True(t, ok, "Expected URL to be found")
	assert.Equal(t, expectedURL, urlRow.URL, "Returned URL should match expected")
	assert.Equal(t, expectedIsDeleted, urlRow.IsDeleted, "Expected is_deleted flag should match")
	assert.NoError(t, mock.ExpectationsWereMet(), "There should be no unfulfilled expectations")
}

func TestURLStorage_GetShortURL(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	ctx := context.Background()
	storage := &URLStorage{conn: mock, ctx: ctx}

	fullURL := "http://example.com"
	expectedShortURL := "short.ly/xyz"

	// Задаем ожидание для SQL запроса
	mock.ExpectQuery(`SELECT short_url FROM urls WHERE url = \$1`).
		WithArgs(fullURL).
		WillReturnRows(pgxmock.NewRows([]string{"short_url"}).
			AddRow(expectedShortURL))

	// Выполнение теста
	actualShortURL, err := storage.GetShortURL(fullURL)

	// Проверка результатов
	assert.NoError(t, err, "Expected no error")
	assert.Equal(t, expectedShortURL, actualShortURL, "Returned short URL should match expected")
	assert.NoError(t, mock.ExpectationsWereMet(), "There should be no unfulfilled expectations")
}

func TestURLStorage_Save(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	ctx := context.Background()
	storage := &URLStorage{conn: mock, ctx: ctx}

	shortURL := "short.ly/xyz"
	fullURL := "http://example.com"
	userID := "user123"

	mock.ExpectExec(`INSERT INTO urls \(short_url, url, user_id\) VALUES \(\$1, \$2, \$3\)`).
		WithArgs(shortURL, fullURL, userID).
		WillReturnResult(pgxmock.NewResult("1", 1)) // 1 строка успешно вставлена

	// Выполнение теста
	err = storage.Save(shortURL, fullURL, userID)

	// Проверка результатов
	assert.NoError(t, err, "Expected no error during save")
	assert.NoError(t, mock.ExpectationsWereMet(), "There should be no unfulfilled expectations")
}

func TestURLStorage_GetURLCount(t *testing.T) {
	// У этой функции нет реальной реализации, но мы можем протестировать, что она возвращает 0
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	ctx := context.Background()
	storage := &URLStorage{conn: mock, ctx: ctx}

	count, _ := storage.GetURLCount()
	assert.Equal(t, 0, count, "Expected URL count to be 0")
}

func TestURLStorage_Ping(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.MonitorPingsOption(true))
	assert.NoError(t, err)
	defer mock.Close()

	ctx := context.Background()
	storage := &URLStorage{conn: mock, ctx: ctx}

	// Задаем ожидание для SQL запроса
	mock.ExpectPing().WillReturnError(nil)

	// Выполнение теста
	assert.True(t, storage.Ping(), "Expected ping to return true")
	assert.NoError(t, mock.ExpectationsWereMet(), "There should be no unfulfilled expectations")
}

func TestURLStorage_Close(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)

	ctx := context.Background()
	storage := &URLStorage{conn: mock, ctx: ctx}

	// Ожидаем, что Close будет вызван на соединении
	mock.ExpectClose()

	// Выполнение теста
	storage.Close()
}
