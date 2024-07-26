package storage

type DataStorageInterface interface {
	Save(row DataStorageRow)
	LoadData() []DataStorageRow
}
