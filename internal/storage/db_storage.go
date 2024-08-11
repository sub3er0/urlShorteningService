package storage

import (
	"context"
	"github.com/jackc/pgx/v4"
	"log"
)

type PgStorage struct {
	conn *pgx.Conn
	ctx  context.Context
}

func (pgs *PgStorage) Init(connectionString string) {
	pgs.ctx = context.Background()
	var err error
	pgs.conn, err = pgx.Connect(pgs.ctx, connectionString)
	if err != nil {
		log.Fatalf("Error while initializing db connection: %v", err)
	}
}

func (pgs *PgStorage) Ping() bool {
	if err := pgs.conn.Ping(pgs.ctx); err != nil {
		return false
	}

	return true
}
func (pgs *PgStorage) Save(row DataStorageRow) error {
	return nil
}
func (pgs *PgStorage) LoadData() ([]DataStorageRow, error) {
	return nil, nil
}

func (pgs *PgStorage) Close() {
	pgs.conn.Close(pgs.ctx)
}
