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
	ID             int           `gorm:"primaryKey;column:id"`
	Theme          string        `gorm:"column:theme"`        // тема заявки
	Description    string        `gorm:"column:description"`  // описание
	WorldStatus    string        `gorm:"column:world_status"` // draft, formed, completed, rejected, deleted
	CreatedAt      time.Time     `gorm:"column:created_at"`
	CreatorID      int           `gorm:"column:creator_id"`
	FormationDate  *time.Time    `gorm:"column:formation_date"`  // дата формирования
	CompletionDate *time.Time    `gorm:"column:completion_date"` // дата завершения
	ModeratorID    *int          `gorm:"column:moderator_id"`
	TotalCost      *float64      `gorm:"column:total_cost"` // вычисляется при завершении
	Planets        []WorldPlanet `gorm:"foreignKey:WorldID;references:ID"`
}

// -------------------------
// WORLD_PLANETS (m-m)
// -------------------------
type WorldPlanet struct {
	//ID        int     `gorm:"primaryKey;column:id"`
	WorldID  int     `gorm:"column:world_id"`
	PlanetID int     `gorm:"column:planet_id"`
	Angle    float64 `gorm:"column:angle"`    // угол
	Distance float64 `gorm:"column:distance"` // радиус/дистанция
	Comment  string  `gorm:"column:comment"`
	Planet   Planet  `gorm:"foreignKey:PlanetID;references:ID"`
}

// -------------------------
// USERS
// -------------------------
type User struct {
	ID       int    `gorm:"primaryKey;column:id"`
	Username string `gorm:"column:username"`
	Password string `gorm:"column:password"`                 // хранить хэш!
	Role     string `gorm:"column:role;default:'astronaut'"` // guest/astronaut/mission_control
}
