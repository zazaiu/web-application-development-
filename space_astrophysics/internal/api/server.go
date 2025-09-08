package api

import (
	"log"
	"net/http"
	"time"

	"space_astrophysics/internal/app/handler"
	"space_astrophysics/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var lines = []string{"first line", "second line", "third line", "fourth line"}

func HelloHandler(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"time":   time.Now().Format("15:04:05"),
		"massiv": lines,
	})
}
func StartServer() {
	log.Println("Starting server")
	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория")
	}

	handler := handler.NewHandler(repo)

	r := gin.Default()

	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/Hello", handler.GetOrders)
	r.GET("/order/:id", handler.GetOrder) // вот наш новый обработчик

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	log.Println("Server down")
}
