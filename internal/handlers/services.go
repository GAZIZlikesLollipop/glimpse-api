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
	"strings"

	"github.com/gin-gonic/gin"
)

func DeleteAllMessagesService(c *gin.Context) {
	var req internal.MsgServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("Ошибка обработки body: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки body"})
		return
	}
	if req.FriendId != nil {
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
			Delete(&internal.Message{}).Error; err != nil {
			log.Println("Ошибка удаления сообщений: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления сообщений"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "Успешное удаление сообщений!"})
}

func deleteMessagesRequest(request internal.MsgServiceRequest) error {
	jsonBody, err := json.Marshal(&request)
	if err != nil {
		log.Println("Ошибка создания тела запроса:", err)
		return err
	}
	req, err := http.NewRequest("POST", "http://msg-service:8080/messages/service", bytes.NewBuffer(jsonBody))
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
		if os.IsExist(err) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления аватарки"})
			log.Println("Ошибка удаления аватарки: ", err)
			return
		}
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

func getUserDataRequest(userId int64) (*internal.User, error) {
	req, err := http.NewRequest("GET", fmt.Sprint("http://user-service:8080/users/", userId), strings.NewReader(""))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("ошибка получения ответа")
	}
	var user internal.User
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, err
	}
	return &user, nil
}
