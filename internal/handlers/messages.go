package handlers

import (
	"api/internal"
	"api/utils"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func AddSentMessage(c *gin.Context) {
	rawReceiverId := c.Param("receiverId")
	receiverId, err := strconv.ParseInt(rawReceiverId, 10, 64)
	if err != nil {
		log.Println("Ошибка конвретации айди получателя: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка конвретации айди получателя"})
		return
	}
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
	var sendMessage internal.Message
	if err := c.ShouldBindJSON(&sendMessage); err != nil {
		log.Println("Ошибка обработки тела: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки тела"})
		return
	}
	sId := uint64(userId)
	rId := uint64(receiverId)
	sendMessage.SenderId = &sId
	sendMessage.ReceiverId = &rId

	if err := utils.Db.Create(&sendMessage).Error; err != nil {
		log.Println("Ошибка добавления сообщения: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка добавления сообщения"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Вы успшено добавили сообщение"})
}

func DeleteSentMessage(c *gin.Context) {
	id := c.Param("id")
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

	if err := utils.Db.Where("sender_id = ? AND id = ?", userId, id).Delete(&internal.Message{}).Error; err != nil {
		log.Println("Ошибка Получения сообщения: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка Получения сообщения"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Успшеное удаление сообщения"})
}

func UpdateSentMessage(c *gin.Context) {
	var updateSentMessage, sentMessage internal.Message
	id := c.Param("id")

	if err := c.ShouldBindJSON(&updateSentMessage); err != nil {
		log.Println("Ошибка чтения тела запроса: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка чтения тела запроса"})
		return
	}

	if err := utils.Db.Where("id = ?", id).First(&sentMessage).Error; err != nil {
		log.Println("Ошибка получения сообщения: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получния сообщения"})
		return
	}

	if strings.TrimSpace(updateSentMessage.Content) != "" {
		sentMessage.Content = updateSentMessage.Content
	}

	if updateSentMessage.IsChecked != sentMessage.IsChecked {
		sentMessage.IsChecked = updateSentMessage.IsChecked
	}

	if err := utils.Db.Save(&sentMessage).Error; err != nil {
		log.Println("Ошибка обнволения сообщения: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления сообщения"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Успешное обновление сообщения"})
}

func DeleteReceivedMessage(c *gin.Context) {
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

	id := c.Param("id")
	if err := utils.Db.Where("receiver_id = ? AND id = ?", userId, id).Delete(&internal.Message{}).Error; err != nil {
		log.Println("Ошибка удаления сообщения: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления сообщения"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Успшеное удаление сообщения"})
}
