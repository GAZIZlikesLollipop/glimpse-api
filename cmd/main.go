package main

import (
	"log"
	"strings"

	"api/utils"

	"api/internal"

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
	protected := r.Group("/api")
	protected.Use(internal.AuthMiddleWare())
	{
		protected.GET("/user", internal.GetUser)
		protected.DELETE("/user", internal.DeleteUser)
		protected.PATCH("/user", internal.UpdateUser)
	}
	r.POST("/signUp", internal.SignUp)
	r.POST("/signIn", internal.SignIn)
	if err := r.RunTLS("0.0.0.0:8080", "server.crt", "server.key"); err != nil {
		log.Fatalln("Ошибка запуска сервера: ", err)
		return
	}
}
