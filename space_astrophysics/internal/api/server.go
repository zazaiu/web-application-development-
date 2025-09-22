package api

import (
	"html/template"
	"log"
	"net/http"

	"space_astrophysics/internal/app/handler"
	"space_astrophysics/internal/app/repository"

	"github.com/gin-gonic/gin"
)

func StartServer() {
	repo, err := repository.NewRepository()
	if err != nil {
		log.Fatal(err)
	}
	h := handler.NewHandler(repo)

	r := gin.Default()

	// ✅ доверяем только localhost, убираем warning
	r.SetTrustedProxies([]string{"127.0.0.1"})

	// подключаем шаблоны
	r.SetFuncMap(template.FuncMap{})
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resourse")

	// ==== маршруты ====
	r.GET("/planets", h.ListPlanets)
	r.GET("/planets/:id", h.ShowPlanetDetail)

	r.GET("/world/:id", h.ViewMissionOrder)
	r.GET("/world/add/:id", h.AddToOrder)

	// redirect корня на список планет
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/planets")
	})

	
	r.Run(":8080")
}
