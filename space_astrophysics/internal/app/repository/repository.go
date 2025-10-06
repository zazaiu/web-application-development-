package repository

import (
	"errors"
	"space_astrophysics/internal/app/models"

	"gorm.io/gorm"
)

type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

// -------------------- PLANETS --------------------

func (r *Repository) GetAllPlanets() ([]models.Planet, error) {
	var planets []models.Planet
	if err := r.DB.Find(&planets).Error; err != nil {
		return nil, err
	}
	return planets, nil
}

func (r *Repository) GetPlanetByID(id int) (models.Planet, error) {
	var planet models.Planet
	if err := r.DB.First(&planet, id).Error; err != nil {
		return models.Planet{}, err
	}
	return planet, nil
}

func (r *Repository) GetDraftWorld(userID int) (*models.World, error) {
	var world models.World
	err := r.DB.Preload("Planets.Planet").
		Where("creator_id = ? AND world_status = ?", userID, "draft").
		First(&world).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // черновой заявки нет
		}
		return nil, err
	}

	return &world, nil
}

// -------------------- WORLDS --------------------

func (r *Repository) GetWorldByID(id int) (models.World, error) {
	var world models.World
	if err := r.DB.Preload("Planets.Planet").First(&world, id).Error; err != nil {
		return models.World{}, err
	}
	return world, nil
}

func (r *Repository) CreateWorld(world *models.World) error {
	return r.DB.Create(world).Error
}

func (r *Repository) AddPlanetToWorld(worldID, planetID, quantity int, isMain bool) error {
	// Проверяем, нет ли уже такой записи (составной PK)
	var existing models.WorldPlanet
	err := r.DB.First(&existing, "world_id = ? AND planet_id = ?", worldID, planetID).Error
	if err == nil {
		return errors.New("планета уже добавлена в заявку")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	wp := models.WorldPlanet{
		WorldID:  worldID,
		PlanetID: planetID,
		Quantity: quantity,
		IsMain:   isMain,
	}
	return r.DB.Create(&wp).Error
}

// Логическое удаление заявки через SQL
func (r *Repository) DeleteWorldSQL(worldID int) error {
	return r.DB.Exec("UPDATE worlds SET world_status='deleted' WHERE id = ?", worldID).Error
}

// Обновление статуса заявки через ORM
func (r *Repository) UpdateWorldStatus(worldID int, status string) error {
	return r.DB.Model(&models.World{}).
		Where("id = ?", worldID).
		Update("world_status", status).Error
}
