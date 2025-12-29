package repository

import (
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"space_astrophysics/internal/app/models"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/gorm"
)

type Repository struct {
	DB     *gorm.DB
	Minio  *minio.Client
	Bucket string
}

// Инициализация репозитория
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

// Если используешь MinIO:
func (r *Repository) InitMinio(endpoint, accessKey, secretKey, bucket string, useSSL bool) error {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return err
	}
	r.Minio = minioClient
	r.Bucket = bucket

	// Создание бакета при необходимости
	exists, errBucketExists := minioClient.BucketExists(nil, bucket)
	if errBucketExists != nil {
		return errBucketExists
	}
	if !exists {
		if err := minioClient.MakeBucket(nil, bucket, minio.MakeBucketOptions{}); err != nil {
			return err
		}
	}
	return nil
}

//
// =====================================================
// PLANETS
// =====================================================
//

// Получить все планеты
func (r *Repository) GetAllPlanets() ([]models.Planet, error) {
	var planets []models.Planet
	if err := r.DB.Find(&planets).Error; err != nil {
		return nil, err
	}
	return planets, nil
}

func (r *Repository) GetWorldsByCreator(userID int) ([]models.World, error) {
	var worlds []models.World
	if err := r.DB.
		Where("creator_id = ?", userID).
		Find(&worlds).Error; err != nil {
		return nil, err
	}
	return worlds, nil
}

// Получить планету по ID
func (r *Repository) GetPlanetByID(id int) (models.Planet, error) {
	var planet models.Planet
	if err := r.DB.First(&planet, id).Error; err != nil {
		return models.Planet{}, err
	}
	return planet, nil
}

// Создать новую планету
func (r *Repository) CreatePlanet(p *models.Planet) error {
	return r.DB.Create(p).Error
}

// Обновить планету
func (r *Repository) UpdatePlanet(id int, update models.Planet) error {
	return r.DB.Model(&models.Planet{}).Where("id = ?", id).Updates(update).Error
}

// Удалить планету
func (r *Repository) DeletePlanet(id int) error {
	return r.DB.Delete(&models.Planet{}, id).Error
}

// Загрузить изображение планеты в MinIO
func (r *Repository) UploadPlanetImageToMinio(planetID int, file *multipart.FileHeader) (string, error) {
	if r.Minio == nil {
		return "", errors.New("MinIO не инициализирован")
	}
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	objectName := fmt.Sprintf("planet_%d%s", planetID, filepath.Ext(file.Filename))
	_, err = r.Minio.PutObject(
		nil,
		r.Bucket,
		objectName,
		src,
		file.Size,
		minio.PutObjectOptions{ContentType: file.Header.Get("Content-Type")},
	)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://%s/%s/%s", r.Minio.EndpointURL().Host, r.Bucket, objectName)
	// обновим поле image_url
	r.DB.Model(&models.Planet{}).Where("id = ?", planetID).Update("image_url", url)
	return url, nil
}

//
// =====================================================
// WORLDS
// =====================================================
//

func (r *Repository) GetDraftWorld(userID int) (*models.World, error) {
	var world models.World
	err := r.DB.Preload("Planets").
		Where("creator_id = ? AND world_status = ?", userID, "draft").
		First(&world).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &world, nil
}

func (r *Repository) GetWorldByID(id int) (models.World, error) {
	var world models.World
	if err := r.DB.Preload("Planets").First(&world, id).Error; err != nil {
		return models.World{}, err
	}
	return world, nil
}

// CreateWorld создает новую заявку с использованием RAW SQL для избежания ON CONFLICT
func (r *Repository) CreateWorld(world *models.World) error {
	// Используем raw SQL для вставки, чтобы избежать проблем с ON CONFLICT
	sql := `INSERT INTO worlds (theme, description, world_status, created_at, creator_id) 
            VALUES (?, ?, ?, ?, ?) RETURNING id`

	return r.DB.Raw(sql,
		world.Theme,
		world.Description,
		world.WorldStatus,
		world.CreatedAt,
		world.CreatorID).Scan(&world.ID).Error
}

// AddPlanetToWorld добавляет планету в заявку (если ещё нет).
// Теперь сохраняет angle, distance и comment вместо quantity/is_main.
func (r *Repository) AddPlanetToWorld(worldID, planetID int, angle, distance float64, comment string) error {
	var existing models.WorldPlanet
	// корректный вызов First с условием и параметрами:
	err := r.DB.Where("world_id = ? AND planet_id = ?", worldID, planetID).First(&existing).Error
	if err == nil {
		return errors.New("планета уже добавлена в заявку")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	wp := models.WorldPlanet{
		WorldID:  worldID,
		PlanetID: planetID,
		Angle:    angle,
		Distance: distance,
		Comment:  comment,
	}
	return r.DB.Create(&wp).Error
}

func (r *Repository) DeleteWorldSQL(worldID int) error {
	return r.DB.Exec("UPDATE worlds SET world_status='deleted' WHERE id = ?", worldID).Error
}

func (r *Repository) UpdateWorldFields(id int, update map[string]interface{}) error {
	return r.DB.Model(&models.World{}).Where("id = ?", id).Updates(update).Error
}

func (r *Repository) GetWorldsFiltered(status, from, to string) ([]models.World, error) {
	query := r.DB.Preload("Planets")
	if status != "" {
		query = query.Where("world_status = ?", status)
	}
	if from != "" {
		query = query.Where("created_at >= ?", from)
	}
	if to != "" {
		query = query.Where("created_at <= ?", to)
	}
	var worlds []models.World
	if err := query.Find(&worlds).Error; err != nil {
		return nil, err
	}
	return worlds, nil
}

func (r *Repository) FormWorld(id int) error {
	return r.DB.Model(&models.World{}).
		Where("id = ?", id).
		Update("world_status", "formed").Error
}

// CompleteWorld завершает заявку с расчетом астрономических параметров
func (r *Repository) CompleteWorld(id int, totalDistance, averageAngle float64) error {
	return r.DB.Model(&models.World{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"world_status": "completed",
			"total_cost":   totalDistance, // Используем существующее поле total_cost для хранения расстояния
			"completed_at": time.Now(),
		}).Error
}

//
// =====================================================
// WORLD-PLANET связи
// =====================================================
//

func (r *Repository) UpdateWorldPlanet(worldID, planetID, quantity int, isMain bool) error {
	return r.DB.Model(&models.WorldPlanet{}).
		Where("world_id = ? AND planet_id = ?", worldID, planetID).
		Updates(map[string]interface{}{
			"quantity": quantity,
			"is_main":  isMain,
		}).Error
}

func (r *Repository) GetWorldPlanet(worldID, planetID int) (*models.WorldPlanet, error) {
	var wp models.WorldPlanet
	if err := r.DB.Where("world_id = ? AND planet_id = ?", worldID, planetID).First(&wp).Error; err != nil {
		return nil, err
	}
	return &wp, nil
}

// Обновить данные по углу, дистанции и комментарию
func (r *Repository) UpdateWorldPlanetFields(worldID, planetID int, angle, distance float64, comment string) error {
	return r.DB.Model(&models.WorldPlanet{}).
		Where("world_id = ? AND planet_id = ?", worldID, planetID).
		Updates(map[string]interface{}{
			"angle":    angle,
			"distance": distance,
			"comment":  comment,
		}).Error
}

// Удалить связь мир–планета
func (r *Repository) DeleteWorldPlanet(worldID, planetID int) error {
	return r.DB.Where("world_id = ? AND planet_id = ?", worldID, planetID).
		Delete(&models.WorldPlanet{}).Error
}

//
// =====================================================
// USERS
// =====================================================
//

func (r *Repository) CreateUser(u *models.User) error {
	return r.DB.Create(u).Error
}

func (r *Repository) GetUserByID(id int) (models.User, error) {
	var user models.User
	if err := r.DB.First(&user, id).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *Repository) UpdateUser(id int, update models.User) error {
	return r.DB.Model(&models.User{}).Where("id = ?", id).Updates(update).Error
}

func (r *Repository) Authenticate(username, password string) (models.User, error) {
	var user models.User
	if err := r.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return models.User{}, err
	}
	if strings.TrimSpace(user.Password) != strings.TrimSpace(password) {
		return models.User{}, errors.New("неверный пароль")
	}
	return user, nil
}

// /////////////////////
// Простое обновление пути изображения планеты (если не используется MinIO)
func (r *Repository) UpdatePlanetImage(planetID int, imagePath string) error {
	return r.DB.Model(&models.Planet{}).
		Where("id = ?", planetID).
		Update("image_url", imagePath).Error
}

// Получить или создать черновик заявки - также обновляем с RAW SQL
func (r *Repository) GetOrCreateDraftWorld(userID int) (*models.World, error) {
	world, err := r.GetDraftWorld(userID)
	if err != nil {
		return nil, err
	}
	if world != nil {
		return world, nil
	}

	newWorld := models.World{
		CreatorID:   userID,
		WorldStatus: "draft",
		CreatedAt:   time.Now(),
		Theme:       "Новая заявка", // Добавляем тему по умолчанию
	}

	// Используем RAW SQL для создания
	sql := `INSERT INTO worlds (theme, description, world_status, created_at, creator_id) 
            VALUES (?, ?, ?, ?, ?) RETURNING id`

	if err := r.DB.Raw(sql,
		newWorld.Theme,
		newWorld.Description,
		newWorld.WorldStatus,
		newWorld.CreatedAt,
		newWorld.CreatorID).Scan(&newWorld.ID).Error; err != nil {
		return nil, err
	}

	return &newWorld, nil
}

// Подсчёт планет в черновике пользователя
func (r *Repository) GetUserCartCount(userID int) (int, error) {
	world, err := r.GetDraftWorld(userID)
	if err != nil {
		return 0, err
	}
	if world == nil {
		return 0, nil
	}

	var count int64
	err = r.DB.Model(&models.WorldPlanet{}).
		Where("world_id = ?", world.ID).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}
