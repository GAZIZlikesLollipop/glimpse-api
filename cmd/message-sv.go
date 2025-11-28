package main

import (
	"api/internal"
	"api/internal/handlers"
	"api/utils"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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
	utils.MsgDb, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalln("Ошибка инциализации базы данных: ", err)
		return
	}
	if err := utils.MsgDb.AutoMigrate(&internal.Message{}); err != nil {
		log.Fatalln("Ошибка миграции таблиц: ", err)
		return
	}
	utils.TcpCns = make(map[string]*websocket.Conn)
	r := gin.Default()
	protected := r.Group("/msg")
	protected.Use(utils.AuthMiddleWare())
	{
		protected.POST("/messages/:receiverId", handlers.AddMessage)
		protected.DELETE("/messages/:id", handlers.DeleteMessage)
		protected.PATCH("/messages/:id", handlers.UpdateMessage)

		protected.GET("/messages/:receiverId", handlers.GetChatMessages)
		protected.GET("/ws", handlers.WebSocket)
	}

	r.POST("/messages/service", handlers.DeleteAllMessagesService)

	if err := r.Run("0.0.0.0:8080"); err != nil {
		log.Fatalln("Ошибка запуска сервиса сообщений: ", err)
		return
	}
}
