package main

import (
	"log"
	"path/filepath"
	"strings"

	"api/utils"

	"api/internal"
	"api/internal/handlers"

	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	if err := strings.TrimSpace(os.Getenv("JWT_SECRET_KEY")); err == "" {
		if err := utils.GenerateJWTSecretKey(); err != nil {
			log.Fatalln("Ошибка генерации jwtSecretKey: ", err)
			return
		}
	}
	var err error
	r := gin.Default()
	dsn := "host=localhost user=postgres dbname=glimpsedb port=5432 sslmode=disable password=1234"
	utils.Db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalln("Ошибка инциализации базы данных: ", err)
		return
	}
	if err := utils.Db.AutoMigrate(&internal.User{}, &internal.Message{}); err != nil {
		log.Fatalln("Ошибка миграции таблиц: ", err)
		return
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("Ошибка поулчения домашней диреткории: ", err)
	}
	absolutePath := filepath.Join(homeDir, "glimpse-media")
	if err := os.MkdirAll(absolutePath, 0755); err != nil {
		log.Fatalln("Ошибка создания диреткории: ", err)
	}
	r.Static("/glimpse-media", absolutePath)
	protected := r.Group("/api")
	protected.Use(handlers.AuthMiddleWare())
	{
		protected.GET("/users", handlers.GetUser)
		protected.DELETE("/users", handlers.DeleteUser)
		protected.PATCH("/users", handlers.UpdateUser)

		protected.GET("/messages/sent", handlers.GetSentMessages)
		protected.POST("/messages/sent/:receiverId", handlers.AddSentMessage)
		protected.DELETE("/messages/sent/:id", handlers.DeleteSentMessage)
		protected.PATCH("/messages/sent/:id", handlers.UpdateSentMessage)

		protected.GET("/messages/received", handlers.GetReceivedMessages)
		protected.DELETE("/messages/received/:id", handlers.DeleteReceivedMessage)

		protected.GET("/friends/:id", handlers.AddFriend)
		protected.DELETE("/friends/:id", handlers.DeleteFriend)

	}
	r.POST("/signUp", handlers.SignUp)
	r.POST("/signIn", handlers.SignIn)
	if err := r.RunTLS("0.0.0.0:8080", "server.crt", "server.key"); err != nil {
		log.Fatalln("Ошибка запуска сервера: ", err)
		return
	}
}
