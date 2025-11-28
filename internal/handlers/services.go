package handlers

import (
	"api/internal"
	"api/utils"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AddUserMessageService(c *gin.Context) {
	var req internal.UserSeriveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("Ошибка обработки body: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки body"})
		return
	}

	tx := utils.UserDb.Begin()

	updateResult := tx.Model(&internal.User{}).
		Where("id = ?", req.SenderId).
		Update("sent_messages", gorm.Expr("ARRAY_APPEND(sent_messages, ?)", req.MsgId))

	if updateResult.Error != nil {
		tx.Rollback()
		log.Println("Ошибка обновления данных отправителя:", updateResult.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления данных отправителя"})
		return
	}

	if updateResult.RowsAffected == 0 {
		tx.Rollback()
		log.Println("Ошибка: Отправитель с ID", req.SenderId, "не найден.")
		c.JSON(http.StatusNotFound, gin.H{"error": "Отправитель не найден"})
		return
	}

	updateResult = tx.Model(&internal.User{}).
		Where("id = ?", req.ReceiverId).
		Update("received_messages", gorm.Expr("ARRAY_APPEND(received_messages, ?)", req.MsgId))

	if updateResult.Error != nil {
		tx.Rollback()
		log.Println("Ошибка обновления данных получателя:", updateResult.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления данных получателя"})
		return
	}

	if updateResult.RowsAffected == 0 {
		tx.Rollback()
		log.Println("Ошибка: Получатель с ID", req.ReceiverId, "не найден.")
		c.JSON(http.StatusNotFound, gin.H{"error": "Получатель не найден"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Успешное добовления сообщения"})
}

func addMessageRequest(request internal.UserSeriveRequest) error {
	jsonBody, err := json.Marshal(&request)
	if err != nil {
		log.Println("Ошибка создания тела запроса:", err)
		return err
	}
	req, err := http.NewRequest("POST", "http://user-service:8080/users/service/add", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Println("Ошибка создания запроса: ", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Ошибка отправки запроса: ", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Println("Ошибка получения ответа")
		return errors.New("ошибка получения ответа")
	}
	return nil
}

func DeleteUserMessageService(c *gin.Context) {
	var req internal.UserSeriveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("Ошибка обработки body: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки body"})
		return
	}

	tx := utils.UserDb.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Println("Паника в транзакции: ", r)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка сервера"})
		}
	}()

	var sender internal.User
	if err := tx.First(&sender, req.SenderId).Error; err != nil {
		tx.Rollback()
		log.Println("Ошибка получения данных отправителя: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения данных отправителя"})
		return
	}

	idxSender := slices.Index(sender.SentMessages, req.MsgId)
	if idxSender >= 0 {
		sender.SentMessages = append(sender.SentMessages[:idxSender], sender.SentMessages[idxSender+1:]...)

		if err := tx.Model(&sender).Update("sent_messages", sender.SentMessages).Error; err != nil {
			tx.Rollback()
			log.Println("Ошибка обновления данных отправителя: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления данных отправителя"})
			return
		}
	}

	var receiver internal.User
	if err := tx.First(&receiver, req.ReceiverId).Error; err != nil {
		tx.Rollback()
		log.Println("Ошибка получения данных получателя: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения данных получателя"})
		return
	}

	idxReceiver := slices.Index(receiver.ReceivedMessages, req.MsgId)
	if idxReceiver >= 0 {
		receiver.ReceivedMessages = append(receiver.ReceivedMessages[:idxReceiver], receiver.ReceivedMessages[idxReceiver+1:]...)

		if err := tx.Model(&receiver).Update("received_messages", receiver.ReceivedMessages).Error; err != nil {
			tx.Rollback()
			log.Println("Ошибка обновления данных получателя: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления данных получателя"})
			return
		}
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Успешное удаление сообщения"})
}

func deleteMessageRequest(request internal.UserSeriveRequest) error {
	jsonBody, err := json.Marshal(&request)
	if err != nil {
		log.Println("Ошибка создания тела запроса:", err)
		return err
	}
	req, err := http.NewRequest("POST", "http://user-service:8080/users/service/delete", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Println("Ошибка создания запроса: ", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Ошибка отправки запроса: ", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Println("Ошибка получения ответа")
		return errors.New("ошибка получения ответа")
	}
	return nil
}

func DeleteAllMessagesService(c *gin.Context) {
	var req internal.MsgServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("Ошибка обработки body: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки body"})
		return
	}
	var messages []internal.Message
	if req.FriendId != nil {
		if err := utils.MsgDb.
			Where(
				"sender_id = ? AND receiver_id = ? OR sender_id = ? AND receiver_id = ?",
				req.UserId,
				req.FriendId,
				req.FriendId,
				req.UserId,
			).
			Find(&messages).Error; err != nil {
			log.Println("Ошибка удаления сообщений: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления сообщений"})
			return
		}
		if err := utils.MsgDb.
			Where(
				"sender_id = ? AND receiver_id = ? OR sender_id = ? AND receiver_id = ?",
				req.UserId,
				req.FriendId,
				req.FriendId,
				req.UserId,
			).
			Delete(&internal.Message{}).Error; err != nil {
			log.Println("Ошибка удаления сообщений: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления сообщений"})
			return
		}
	} else {
		if err := utils.MsgDb.
			Where(
				"sender_id = ? OR receiver_id = ?",
				req.UserId,
				req.UserId,
			).
			Find(&messages).Error; err != nil {
			log.Println("Ошибка удаления сообщений: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления сообщений"})
			return
		}
		if err := utils.MsgDb.
			Where(
				"sender_id = ? OR receiver_id = ?",
				req.UserId,
				req.UserId,
			).
			Delete(&internal.Message{}).Error; err != nil {
			log.Println("Ошибка удаления сообщений: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления сообщений"})
			return
		}
	}
	c.JSON(http.StatusOK, messages)
}

func deleteMessagesRequest(request internal.MsgServiceRequest) (*[]int64, error) {
	jsonBody, err := json.Marshal(&request)
	if err != nil {
		log.Println("Ошибка создания тела запроса:", err)
		return nil, err
	}
	req, err := http.NewRequest("POST", "http://msg-service:8080/messages/service", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Println("Ошибка создания запроса: ", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Ошибка отправки запроса: ", err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("Ошибка получения ответа")
		return nil, errors.New("ошибка получения ответа")
	}
	defer resp.Body.Close()
	var (
		msgs    []internal.Message
		msg_ids []int64
	)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Ошибка чтения ответа: ", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &msgs); err != nil {
		log.Println("Ошибка чтения ответа: ", err)
		return nil, err
	}
	for i := range msgs {
		msg_ids = append(msg_ids, msgs[i].Id)
	}
	return &msg_ids, nil
}

func AddMediaService(c *gin.Context) {
	name := c.Param("name")
	if err := utils.SaveAvatarFile(c, name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения файла"})
		log.Println("Ошибка сохранения аватарки: ", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Успешное добавление аватарки"})
}

func addMediaRequest(request internal.MediaServiceRequest) error {
	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)

	pathtofile := filepath.Join("/app", "glimpse", "media", fmt.Sprintf("%s%s", request.UserName, strings.ToLower(filepath.Ext(request.AvatarExt))))
	file, err := os.Open(pathtofile)
	if err != nil {
		fmt.Println("Ошибка открытия файла:", err)
		return err
	}
	defer file.Close()

	part, err := writer.CreateFormFile("avatar", filepath.Base(pathtofile))
	if err != nil {
		return err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("http://media-service:8080/media/%s", request.UserName), body)
	if err != nil {
		fmt.Println("Ошибка создания запроса:", err)
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Ошибка отправки запроса: ", err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("Ошибка получения ответа")
		return errors.New("ошибка получения ответа")
	}
	defer resp.Body.Close()
	return nil
}

func DeleteMediaService(c *gin.Context) {
	bytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка чтения байтов"})
		log.Println("Ошибка чтения байтов: ", err)
		return
	}
	var data internal.MediaServiceRequest
	if err := json.Unmarshal(bytes, &data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка чтения запроса"})
		log.Println("Ошибка чтения запроса: ", err)
		return
	}
	name := c.Param("name")
	if err := os.Remove(filepath.Join("/app", "glimpse", "media", fmt.Sprintf("%s%s", name, strings.ToLower(filepath.Ext(data.AvatarExt))))); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления аватарки"})
		log.Println("Ошибка удаления аватарки: ", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Успешное удаление аватарки"})
}

func deleteMediaRequest(request internal.MediaServiceRequest) error {
	body, err := json.Marshal(request)
	if err != nil {
		log.Println("Ошибка создания тела запроса:", err)
		return err
	}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("http://media-service:8080/media/%s", request.UserName), bytes.NewBuffer(body))
	if err != nil {
		log.Println("Ошибка создания запроса: ", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Ошибка отправки запроса: ", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Println("Ошибка получения ответа")
		return errors.New("ошибка получения ответа")
	}
	return nil
}
