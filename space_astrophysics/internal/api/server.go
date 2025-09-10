package api

import (
	"log"
	"space_astrophysics/internal/app/handler"
	"space_astrophysics/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	log.Println("Starting server")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория")
		return
	}

	h := handler.NewHandler(repo)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/services", h.GetServices)
	r.GET("/services/:id", h.GetService)
	r.GET("/order/:id", h.GetOrder)
	r.GET("/order/add/:id", h.AddToOrder)
	r.GET("/order/clear/:id", h.ClearOrder)
	r.GET("/order/calc/:id", h.CalcOrder)

	r.Run(":8080")
	log.Println("Server down")
}
