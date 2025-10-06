package models

import "time"

// -------------------------
// PLANETS
// -------------------------
type Planet struct {
	ID          int     `gorm:"primaryKey;column:id"`
	Name        string  `gorm:"column:name"`
	Description string  `gorm:"column:description"`
	ImageURL    string  `gorm:"column:image_url"` // nullable
	Status      string  `gorm:"column:status"`    // 'active' / 'deleted'
	A           float64 `gorm:"column:a"`
	E           float64 `gorm:"column:e"`
	Period      float64 `gorm:"column:period"`
	T0          float64 `gorm:"column:t0"`
}

// -------------------------
// WORLDS (Заявки)
// -------------------------
type World struct {
	ID                 int           `gorm:"primaryKey;column:id"`
	WorldStatus        string        `gorm:"column:world_status"` // draft, deleted, formed, completed, rejected
	CreatedAt          time.Time     `gorm:"column:created_at"`
	CreatorID          int           `gorm:"column:creator_id"`
	FormationDate      *time.Time    `gorm:"column:formation_date"`
	CompletionDate     *time.Time    `gorm:"column:completion_date"`
	ModeratorID        *int          `gorm:"column:moderator_id"`
	Planets            []WorldPlanet `gorm:"foreignKey:WorldID;references:ID"`
	Date               string        `gorm:"-"` // дата для расчета
	CalculatedAngle    float64       `gorm:"-"` // динамически рассчитывается
	CalculatedDistance float64       `gorm:"-"`
}

// -------------------------
// WORLD_PLANETS (m-m)
// -------------------------
type WorldPlanet struct {
	WorldID  int    `gorm:"primaryKey;column:world_id;constraint:OnDelete:RESTRICT"`
	PlanetID int    `gorm:"primaryKey;column:planet_id;constraint:OnDelete:RESTRICT"`
	Quantity int    `gorm:"column:quantity;default:1"`
	IsMain   bool   `gorm:"column:is_main;default:false"`
	Planet   Planet `gorm:"foreignKey:PlanetID;references:ID"`
}

// -------------------------
// USERS
// -------------------------
type User struct {
	ID          int    `gorm:"primaryKey;column:id"`
	Username    string `gorm:"column:username"`
	Password    string `gorm:"column:password"` // хранить хэш!
	IsModerator bool   `gorm:"column:is_moderator;default:false"`
}
