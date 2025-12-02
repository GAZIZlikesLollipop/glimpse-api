package main

import (
	"api/internal/handlers"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	absolutePath := filepath.Join("/app", "glimpse", "media")
	if err := os.MkdirAll(absolutePath, 0755); err != nil {
		log.Fatalln("Ошибка создания диреткории: ", err)
	}
	r.Static("/media", absolutePath)
	r.POST("/media/:name", handlers.AddMediaService)
	r.DELETE("/media/:name", handlers.DeleteMediaService)
	if err := r.Run("0.0.0.0:8080"); err != nil {
		log.Fatalln("Ошибка запуска медиа сервиса: ", err)
		return
	}
}
