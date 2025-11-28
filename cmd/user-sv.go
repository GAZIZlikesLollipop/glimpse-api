package main

import (
	"api/internal"
	"api/internal/handlers"
	"api/utils"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	var err error
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_MASK"),
		os.Getenv("POSTGRES_NAME"),
	)
	utils.UserDb, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalln("Ошибка инциализации базы данных: ", err)
		return
	}
	if err := utils.UserDb.AutoMigrate(&internal.User{}); err != nil {
		log.Fatalln("Ошибка миграции таблиц: ", err)
		return
	}
	r := gin.Default()
	protected := r.Group("/usr")
	protected.Use(utils.AuthMiddleWare())
	{
		protected.GET("/users", handlers.GetUser)
		protected.DELETE("/users", handlers.DeleteUser)
		protected.PATCH("/users", handlers.UpdateUser)

		protected.GET("/friends/app/:id", handlers.AddFriend)
		protected.DELETE("/friends/:id", handlers.DeleteFriend)
	}

	r.POST("/users/service/add", handlers.AddUserMessageService)
	r.POST("/users/service/delete", handlers.DeleteUserMessageService)

	r.GET("/users/friends/:id", handlers.GetFriends)
	r.GET("/users", handlers.GetUsers)

	if err := r.Run("0.0.0.0:8080"); err != nil {
		log.Fatalln("Ошибка запуска сервиса пользователей: ", err)
		return
	}
}
