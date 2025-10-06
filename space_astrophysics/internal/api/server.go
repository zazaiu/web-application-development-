package api

import (
	"log"
	"space_astrophysics/internal/app/handler"
	"space_astrophysics/internal/app/models"
	"space_astrophysics/internal/app/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := "host=localhost port=5433 user=astrouser password=1234 dbname=astrodb sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	DB = db

	err = db.AutoMigrate(&models.Planet{}, &models.World{}, &models.WorldPlanet{})
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Database initialized")
}

func StartServer() {
	InitDB()

	repo := repository.NewRepository(DB)
	h := handler.NewHandler(repo)

	r := gin.Default()

	// --- СТАТИКА ---
	r.Static("/static", "./static")

	// --- ШАБЛОНЫ ---
	r.LoadHTMLGlob("templates/*")

	// --- МАРШРУТЫ ---
	r.GET("/planets", h.ListPlanets)
	r.GET("/planets/:id", h.ShowPlanetDetail)
	r.GET("/world/:id", h.ViewWorld)
	r.POST("/world/add-planet", h.AddPlanetToDraftWorld)
	r.POST("/world/delete/:id", h.DeleteWorld) // удаление заявки
	r.POST("/world/form/:id", h.FormWorld)     // оформление заявки
	log.Println("Server running on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
