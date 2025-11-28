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
	r.POST("/auth/signUp", handlers.SignUp)
	r.POST("/auth/signIn", handlers.SignIn)
	if err := r.Run("0.0.0.0:8080"); err != nil {
		log.Fatalln("Ошибка запуска сервиса аутенефикации: ", err)
		return
	}
}
