package storage

type DataStorageRow struct {
	ID       int    `json:"uuid"`
	ShortURL string `json:"short_url"`
	URL      string `json:"original_url"`
}
