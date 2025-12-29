package models

import "time"

// -------------------------
// MEDIA (фото и видео для планет)
// -------------------------
type Media struct {
	ID        int       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	PlanetID  int       `gorm:"column:planet_id;not null" json:"planet_id"`
	FileURL   string    `gorm:"column:file_url;not null" json:"file_url"`   // URL файла в MinIO
	FileType  string    `gorm:"column:file_type;not null" json:"file_type"` // 'image' или 'video'
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	Planet    *Planet   `gorm:"foreignKey:PlanetID;references:ID" json:"planet,omitempty"`
}

func (Media) TableName() string {
	return "media"
}
