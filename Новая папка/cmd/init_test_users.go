package main

import (
	"log"
	"space_astrophysics/internal/app/models"
	"space_astrophysics/internal/app/repository"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Подключаемся к БД
	dsn := "host=localhost port=5433 user=astrouser password=1234 dbname=astrodb sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	// Создаем репозиторий
	repo := repository.NewRepository(db)

	// Создаем обычного пользователя
	user := &models.User{
		Username: "user",
		Password: "1234", // Для демо - простой пароль
		Role:     "astronaut",
	}

	// Создаем модератора
	moderator := &models.User{
		Username: "moderator",
		Password: "1234",
		Role:     "mission_control",
	}

	// Пытаемся создать пользователей
	if err := repo.CreateUser(user); err != nil {
		log.Printf("User already exists or error: %v", err)
	} else {
		log.Printf("Created user: %s (role: %s)", user.Username, user.Role)
	}

	if err := repo.CreateUser(moderator); err != nil {
		log.Printf("Moderator already exists or error: %v", err)
	} else {
		log.Printf("Created moderator: %s (role: %s)", moderator.Username, moderator.Role)
	}

	log.Println("Test users initialization completed!")
	log.Println("User: user / 1234")
	log.Println("Moderator: moderator / 1234")
}
