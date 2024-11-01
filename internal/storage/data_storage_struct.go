package storage

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
