package storage

const tableName = "urls"

// URL представляет структуру таблицы urls.
type URL struct {
	ID        uint   `gorm:"primaryKey"`
	URL       string `gorm:"uniqueIndex;size:100"`
	ShortURL  string `gorm:"uniqueIndex;size:100"`
	UserID    string `gorm:"size:100"`
	IsDeleted bool   `gorm:"default:false"`
}

// UserCookie представляет структуру таблицы users_cookie.
type UserCookie struct {
	ID     uint   `gorm:"primaryKey"`
	UserID string `gorm:"uniqueIndex;size:100"`
}
