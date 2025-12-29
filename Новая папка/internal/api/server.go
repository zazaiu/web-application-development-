package api

import (
	"log"
	"net/http"
	_ "space_astrophysics/internal/api/docs"
	"space_astrophysics/internal/app/handler"
	"space_astrophysics/internal/app/models"
	"space_astrophysics/internal/app/repository"
	"space_astrophysics/internal/utils"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Разрешаем PWA на GitHub Pages
		allowedOrigins := []string{
			"http://localhost:5173",    // Vite dev server
			"https://zazaiu.github.io", // Ваш GitHub Pages
			"http://localhost:8080",    // Go dev server
			"http://192.168.1.69:8080", // Ваш IP для Tauri
			"tauri://localhost",        // Tauri
		}

		// Проверяем origin
		allowOrigin := "*"
		for _, ao := range allowedOrigins {
			if ao == origin {
				allowOrigin = origin
				break
			}
		}

		// Для GitHub Pages также проверяем поддомены
		if origin == "" || origin == "null" {
			// Может быть в Service Worker или других случаях
			allowOrigin = "*"
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Range")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

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
// ИНИЦИАЛИЗАЦИЯ ТЕСТОВЫХ ПОЛЬЗОВАТЕЛЕЙ
// ----------------------
func InitTestUsers(db *gorm.DB) {
	repo := repository.NewRepository(db)

	// Создаем обычного пользователя
	user := &models.User{
		Username: "user",
		Password: "1234",
		Role:     "astronaut",
	}

	// Создаем модератора
	moderator := &models.User{
		Username: "moderator",
		Password: "1234",
		Role:     "mission_control",
	}

	// Создаем пользователей (игнорируем ошибки если уже существуют)
	if err := repo.CreateUser(user); err != nil {
		log.Printf("⚠️ User creation: %v", err)
	} else {
		log.Printf("✅ Created user: %s", user.Username)
	}

	if err := repo.CreateUser(moderator); err != nil {
		log.Printf("⚠️ Moderator creation: %v", err)
	} else {
		log.Printf("✅ Created moderator: %s", moderator.Username)
	}

	log.Println("🎉 Test users initialized: user/1234 (astronaut), moderator/1234 (mission_control)")
}

// ----------------------
// ЗАПУСК СЕРВЕРА
// ----------------------
func StartServer() {
	InitDB()

	// Инициализируем тестовых пользователей
	InitTestUsers(DB)

	repo := repository.NewRepository(DB)
	h := handler.NewHandler(repo)

	r := gin.Default()
	r.Use(CORSMiddleware())
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
		api.GET("/cart", h.GetCart)
		// Callback от Django - используем другой путь чтобы не конфликтовать
		api.POST("/async-callback/:world_id", h.CompleteWorldCallback)

		// Media - публичный доступ для просмотра (используем /media/planet/:id вместо /planets/:id/media)
		api.GET("/media/planet/:id", h.GetPlanetMedia)
		api.GET("/media/:id", h.GetMediaByID)
	}

	// === Авторизованные пользователи (любая роль) ===
	auth := api.Group("/")
	auth.Use(h.AuthMiddleware())
	{
		// Личный кабинет
		auth.GET("/users/profile", h.GetUserProfile)
		auth.PUT("/users/profile", h.UpdateUserProfile)
		auth.POST("/users/logout", h.Logout)

		// Заявки - ВСЕ используем :id
		auth.GET("/worlds", h.ListWorldsFiltered)
		auth.GET("/worlds/:id", h.ViewWorld)
		auth.POST("/worlds", h.CreateWorld)
		auth.PUT("/worlds/:id", h.UpdateWorld)
		auth.DELETE("/worlds/:id", h.DeleteWorld)
		auth.GET("/worlds/:id/calculation-status", h.GetCalculationStatus)
		auth.POST("/worlds/:id/calculate", h.CalculateOrbitalData)
		// ИЗМЕНЕНИЕ: используем :id вместо :world_id
		auth.POST("/worlds/:id/planet/:planet_id", h.AddPlanetToWorld)
	}

	// === Только для астронавтов ===
	astronaut := api.Group("/")
	astronaut.Use(h.AuthMiddleware("astronaut"))
	{
		astronaut.PUT("/worlds/:id/form", h.FormWorld)
	}

	// === Только для модератора (mission_control) ===
	admin := api.Group("/")
	admin.Use(h.AuthMiddleware("mission_control"))
	{
		admin.POST("/planets", h.CreatePlanet)
		admin.PUT("/planets/:id", h.UpdatePlanet)
		admin.DELETE("/planets/:id", h.DeletePlanet)
		admin.POST("/planets/:id/image", h.UploadPlanetImage)

		// Асинхронный расчет
		admin.PUT("/worlds/:id/complete", h.CompleteWorld)
		admin.POST("/worlds/:id/start-async", h.StartAsyncCalculation)

		// Работа со связями M-M - используем :id вместо :world_id
		admin.GET("/world-planets/:id/:planet_id", h.GetWorldPlanet)
		admin.DELETE("/world-planets/:id/:planet_id", h.DeleteWorldPlanet)
		admin.PUT("/world-planets/:id/:planet_id", h.UpdateWorldPlanet)

		// Media - управление медиа файлами (только модератор)
		admin.POST("/media/planet/:id", h.AddPlanetMedia)
		admin.PUT("/media/:id", h.UpdateMedia)
		admin.DELETE("/media/:id", h.DeleteMedia)

		// Админка Redis
		admin.GET("/admin/sessions", h.ShowAllSessions)
		admin.GET("/admin/redis-check", h.RedisCheck)
	}

	// -------------------------------
	// ЗАПУСК СЕРВЕРА
	// -------------------------------
	log.Println("🚀 Server running on http://localhost:8080")
	log.Println("📚 Swagger available at http://localhost:8080/swagger/index.html")
	log.Println("🔐 Admin sessions at http://localhost:8080/api/admin/sessions")
	log.Println("👤 Test users: user/1234 (astronaut), moderator/1234 (mission_control)")
	log.Println("🔄 Callback endpoint: POST /api/async-callback/:world_id")

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
