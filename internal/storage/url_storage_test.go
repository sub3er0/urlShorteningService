package storage

import (
	"context"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/mock"
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

// MockDBConnection это мок для интерфейса DBConnectionInterface
type MockDBConnection struct {
	mock.Mock
}

// Query выполняет SQL-запрос и возвращает строки результата.
func (m *MockDBConnection) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	call := m.Called(ctx, sql, args)
	return call.Get(0).(pgx.Rows), call.Error(1)
}

// Exec выполняет SQL-команду и возвращает тег команды.
func (m *MockDBConnection) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	call := m.Called(ctx, sql, args)
	return call.Get(0).(pgconn.CommandTag), call.Error(1)
}

// Ping проверяет состояние подключения к базе данных.
func (m *MockDBConnection) Ping(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

// Close закрывает соединение с базой данных.
func (m *MockDBConnection) Close() {
	m.Called()
}

// SendBatch отправляет пакет запросов в базу данных.
func (m *MockDBConnection) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	call := m.Called(ctx, b)
	return call.Get(0).(pgx.BatchResults)
}

// QueryRow возвращает одну строку результата.
func (m *MockDBConnection) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	call := m.Called(ctx, sql, args)
	return call.Get(0).(pgx.Row) // Обратите внимание, что нужно соответствовать типам возвращаемого значения
}

// MockBatchResults - это мок для интерфейса pgx.BatchResults
type MockBatchResults struct {
	mock.Mock
}

// Exec имитирует выполнение одного запроса в пакетных результатах.
func (m *MockBatchResults) Exec() (pgconn.CommandTag, error) {
	args := m.Called()
	return args.Get(0).(pgconn.CommandTag), args.Error(1)
}

// Query имитирует выполнение запроса в пакетных результатах.
func (m *MockBatchResults) Query() (pgx.Rows, error) {
	args := m.Called()
	return args.Get(0).(pgx.Rows), args.Error(1)
}

// QueryRow имитирует выполнение одного запроса и возврат одной строки результата.
func (m *MockBatchResults) QueryRow() pgx.Row {
	args := m.Called()
	return args.Get(0).(pgx.Row)
}

// QueryFunc имитирует выполнение запроса с использованием функции обратного вызова.
func (m *MockBatchResults) QueryFunc(scans []interface{}, f func(pgx.QueryFuncRow) error) (pgconn.CommandTag, error) {
	args := m.Called(scans, f)
	return args.Get(0).(pgconn.CommandTag), args.Error(1)
}

// Close имитирует завершение пакетной операции.
func (m *MockBatchResults) Close() error {
	return m.Called().Error(0)
}

// TestSaveBatchURL юнит тест метода
func TestSaveBatchURL(t *testing.T) {
	mockConn := new(MockDBConnection)
	mockBatchResults := new(MockBatchResults)

	us := &URLStorage{
		conn: mockConn,
	}

	dataStorageRows := []DataStorageRow{
		{URL: "http://example.com", ShortURL: "short.ly/1", UserID: "user1"},
		{URL: "http://example.org", ShortURL: "short.ly/2", UserID: "user2"},
	}

	mockBatchResults.On("Exec").Return(pgconn.CommandTag("INSERT 0 1"), nil).Times(2)
	mockBatchResults.On("Close").Return(nil)
	batch := &pgx.Batch{}
	batch.Queue("INSERT INTO urls (url, short_url, user_id) VALUES ($1, $2, $3) ON CONFLICT (url, short_url) DO NOTHING",
		dataStorageRows[0].URL, dataStorageRows[0].ShortURL, dataStorageRows[0].UserID)
	batch.Queue("INSERT INTO urls (url, short_url, user_id) VALUES ($1, $2, $3) ON CONFLICT (url, short_url) DO NOTHING",
		dataStorageRows[1].URL, dataStorageRows[1].ShortURL, dataStorageRows[1].UserID)
	mockConn.On("SendBatch", context.Background(), batch).Return(mockBatchResults)
	mockConn.On("Exec").Return(pgconn.CommandTag{}, nil).Times(len(dataStorageRows))

	err := us.SaveBatch(dataStorageRows)
	assert.NoError(t, err)
}

// Connect имитация метода
func (m *MockDBConnection) Connect(ctx context.Context, dsn string) (DBConnectionInterface, error) {
	args := m.Called(ctx, dsn)
	return args.Get(0).(DBConnectionInterface), args.Error(1)
}

// TestInit юнит тест метода
func TestInit(t *testing.T) {
	mockConn := new(MockDBConnection)
	us := &URLStorage{}

	mockConn.On("Connect", mock.Anything, mock.Anything).Return(mockConn, nil)

	err := us.Init("postgres://postgres:326717@localhost:5432/shortener?sslmode=disable")
	us.Close()
	assert.NoError(t, err)
}
