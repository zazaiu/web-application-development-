package api

import (
	"log"
	"net/http"
	_ "space_astrophysics/internal/api/docs"
	"space_astrophysics/internal/app/handler"
	"space_astrophysics/internal/app/repository"
	"space_astrophysics/internal/utils"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// ----------------------
// ИНИЦИАЛИЗАЦИЯ БД
// ----------------------
func InitDB() {
	dsn := "host=localhost port=5433 user=astrouser password=1234 dbname=astrodb sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Ошибка подключения к БД: %v", err)
	}

	DB = db
	log.Println("✅ Database initialized")
}

// ----------------------
// ЗАПУСК СЕРВЕРА
// ----------------------
func StartServer() {
	InitDB()

	repo := repository.NewRepository(DB)
	h := handler.NewHandler(repo)

	r := gin.Default()
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")
	utils.InitRedis("localhost:6379", "", 0)
	// -------------------------------
	// HTML фронт
	// -------------------------------
	r.GET("/planets", func(ctx *gin.Context) {
		planets, err := h.Repo.GetAllPlanets()
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Ошибка: %v", err)
			return
		}
		ctx.HTML(http.StatusOK, "service_list.html", gin.H{
			"planets": planets,
		})
	})

	// -------------------------------
	// Swagger UI
	// -------------------------------
	//	docs.SwaggerInfo.BasePath = "/api"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// -------------------------------
	// REST API
	// -------------------------------

	// === ГОСТЬ (без токена) ===
	api := r.Group("/api")
	{
		api.POST("/users/register", h.RegisterUser)
		api.POST("/users/login", h.Login)
		api.GET("/planets", h.ListPlanets)
		api.GET("/planets/:id", h.ShowPlanetDetail)
	}
	// === Авторизованные пользователи (astronaut, mission_control) ===
	auth := api.Group("/")
	auth.Use(h.AuthMiddleware("astronaut", "mission_control"))
	{
		// Заявки
		auth.GET("/worlds", h.ListWorldsFiltered)
		auth.GET("/worlds/:id", h.ViewWorld)
		auth.PUT("/worlds/:id", h.UpdateWorld)
		auth.PUT("/worlds/:id/form", h.FormWorld)
		auth.DELETE("/worlds/:id", h.DeleteWorld)

		// Личный кабинет
		auth.GET("/users/me", h.GetUserProfile)
		auth.PUT("/users/me", h.UpdateUserProfile)
		auth.POST("/users/logout", h.Logout)

		// Корзина
		auth.GET("/cart", h.GetCartIcon)
		auth.POST("/planets/:id/add-to-world", h.AddPlanetToWorld)
	}

	// === Только для модератора (mission_control) ===
	admin := api.Group("/")
	admin.Use(h.AuthMiddleware("mission_control"))
	{
		admin.POST("/planets", h.CreatePlanet)
		admin.PUT("/planets/:id", h.UpdatePlanet)
		admin.DELETE("/planets/:id", h.DeletePlanet)
		admin.POST("/planets/:id/image", h.UploadPlanetImage)

		// Завершение заявок
		admin.PUT("/worlds/:id/complete", h.CompleteWorld)

		// Работа с связями M-M
		wp := admin.Group("/world-planets")
		{
			wp.GET("/:world_id/:planet_id", h.GetWorldPlanet)
			wp.DELETE("/:world_id/:planet_id", h.DeleteWorldPlanet)
			wp.PUT("/:world_id/:planet_id", h.UpdateWorldPlanet)
		}
	}

	// -------------------------------
	// ЗАПУСК СЕРВЕРА
	// -------------------------------
	log.Println("🚀 Server running on http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
