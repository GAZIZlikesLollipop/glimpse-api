package handlers

import (
	"api/internal"
	"api/utils"
	"log"
	"net/http"
	"slices"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetFriends(c *gin.Context) {
	id := c.Param("id")
	var friend internal.User
	if err := utils.UserDb.Preload("Friends").First(&friend, id).Error; err != nil {
		log.Println("Ошибка получения друзей: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения друзей"})
		return
	}
	c.JSON(http.StatusOK, friend.Friends)
}

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
	if err := utils.UserDb.Preload("Friends").First(&user, userId).Error; err != nil {
		log.Println("Ошибка получения пользовтеля: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получеения пользователя"})
		return
	}

	var friend internal.User
	friendId := c.Param("id")
	if err := utils.UserDb.Preload("Friends").First(&friend, friendId).Error; err != nil {
		log.Println("Ошибка получения пользователя: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения пользовaтеля"})
		return
	}

	if err := utils.UserDb.Model(&user).Association("Friends").Append(&friend); err != nil {
		log.Println("Ошибка добавления друга: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка добавления друга"})
		return
	}
	if err := utils.UserDb.Model(&friend).Association("Friends").Append(&user); err != nil {
		log.Println("Ошибка добавления друга: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка добавления друга"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Друг успешно добавлен"})
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

	if err := utils.UserDb.Preload("Friends").First(&user, userId).Error; err != nil {
		log.Println("Ошибка получения пользовтеля: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получеения пользователя"})
		return
	}
	var friend internal.User
	friendId := c.Param("id")
	if err := utils.UserDb.Preload("Friends").First(&friend, friendId).Error; err != nil {
		log.Println("Ошибка получения пользователя: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения пользовaтеля"})
		return
	}
	tx := utils.UserDb.Begin()
	if err := tx.Model(&user).Association("Friends").Delete(&friend); err != nil {
		tx.Rollback()
		log.Println("Ошибка удаления друга: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления друга"})
		return
	}
	if err := tx.Model(&friend).Association("Friends").Delete(&user); err != nil {
		tx.Rollback()
		log.Println("Ошибка удаления друга: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления друга"})
		return
	}
	friendID, err := strconv.ParseInt(friendId, 10, 0)
	if err != nil {
		log.Println("Ошибка преобразования id: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка преобразования id"})
		return
	}
	msgReq := internal.MsgServiceRequest{
		UserId:   userId,
		FriendId: &friendID,
	}
	msgs, err := deleteMessagesRequest(msgReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления сообщений"})
		return
	}
	mesgs := *msgs
	for i := range mesgs {
		user.SentMessages = append(user.SentMessages[:slices.Index(user.SentMessages, mesgs[i])], user.SentMessages[slices.Index(user.SentMessages, mesgs[i])+1:]...)
		user.ReceivedMessages = append(user.ReceivedMessages[:slices.Index(user.ReceivedMessages, mesgs[i])], user.ReceivedMessages[slices.Index(user.ReceivedMessages, mesgs[i])+1:]...)
	}
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления сообщений друга"})
		return
	}
	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Успешноe удаления друга"})
}
