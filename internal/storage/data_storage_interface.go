package storage

type DataStorageInterface interface {
	Save(row DataStorageRow) error
	LoadData() ([]DataStorageRow, error)
}
