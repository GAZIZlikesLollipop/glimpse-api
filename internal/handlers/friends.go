package handlers

import (
	"api/internal"
	"api/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AddFriend(c *gin.Context) {
	rawUserId, exists := c.Get("userId")
	if !exists {
		log.Println("Ошибка получения данных с токена")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения данных с токена"})
		return
	}

	userId, ok := rawUserId.(int64)
	if !ok {
		log.Println("Ошибка преобрзаования айди")
		c.JSON(http.StatusInternalServerError, map[string]any{"error": "Ошибка преобразования айди"})
		return
	}

	var user internal.User

	if err := utils.Db.Preload("SentMessages").Preload("ReceivedMessages").Preload("Friends").First(&user, userId).Error; err != nil {
		log.Println("Ошибка получения пользовтеля: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получеения пользователя"})
		return
	}
	var friend internal.User
	friendId := c.Param("id")
	if err := utils.Db.Preload("Friends").First(&friend, friendId).Error; err != nil {
		log.Println("Ошибка получения пользователя: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения пользовaтеля"})
		return
	}
	if err := utils.Db.Model(&user).Association("Friends").Append(&friend); err != nil {
		log.Println("Ошибка добавления друга: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка добавления друга"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Друг успшено добавлен"})
}

func DeleteFriend(c *gin.Context) {
	rawUserId, exists := c.Get("userId")
	if !exists {
		log.Println("Ошибка получения данных с токена")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения данных с токена"})
		return
	}

	userId, ok := rawUserId.(int64)
	if !ok {
		log.Println("Ошибка преобрзаования айди")
		c.JSON(http.StatusInternalServerError, map[string]any{"error": "Ошибка преобразования айди"})
		return
	}

	var user internal.User

	if err := utils.Db.Preload("SentMessages").Preload("ReceivedMessages").Preload("Friends").First(&user, userId).Error; err != nil {
		log.Println("Ошибка получения пользовтеля: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получеения пользователя"})
		return
	}
	var friend internal.User
	friendId := c.Param("id")
	if err := utils.Db.Preload("Friends").First(&friend, friendId).Error; err != nil {
		log.Println("Ошибка получения пользователя: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения пользовaтеля"})
		return
	}
	if err := utils.Db.Model(&user).Association("Freinds").Delete(&friend); err != nil {
		log.Println("Ошибка удаления друга: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления друга"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Успешноe удаления друга"})
}
