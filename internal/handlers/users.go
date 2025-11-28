package handlers

import (
	"api/internal"
	"api/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
)

func GetUser(c *gin.Context) {
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

	if err := utils.UserDb.Preload("Friends.Friends").First(&user, userId).Error; err != nil {
		log.Println("Ошибка получения пользовтеля: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получеения пользователя"})
		return
	}

	c.JSON(http.StatusOK, user)

}

func SignUp(c *gin.Context) {
	var user internal.User
	name := c.PostForm("name")
	if name == "" {
		log.Println("Клиент не ввеели имя")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Вы не ввели имя"})
		return
	} else {
		user.Name = name
	}
	var count int64
	if err := utils.UserDb.Model(&internal.User{}).Where("name = ?", user.Name).Count(&count).Error; err != nil {
		log.Println("Ошибка получения пользователя: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения пользователя"})
		return
	}

	if count < 1 {
		strPassword := c.PostForm("password")
		if strPassword == "" {
			log.Println("Пользовтель не ввел пароль")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Вы  не ввели пароль"})
			return
		} else {
			password, err := bcrypt.GenerateFromPassword([]byte(strPassword), bcrypt.DefaultCost)
			if err != nil {
				log.Println("Ошибка генерации пароля: ", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генрации пароля"})
				return
			}
			user.Password = string(password)
		}
		bio := c.PostForm("bio")
		if bio != "" {
			user.Bio = bio
		}
		strLatitude := c.PostForm("latitude")
		if strings.TrimSpace(strLatitude) != "" {
			latitude, err := strconv.ParseFloat(strLatitude, 64)
			if err != nil {
				log.Println("Ошибка парсинга широты: ", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка парсинга широты"})
				return
			}
			user.Latitude = latitude
		}
		strLongitude := c.PostForm("longitude")
		if strings.TrimSpace(strLongitude) != "" {
			longitude, err := strconv.ParseFloat(strLongitude, 64)
			if err != nil {
				log.Println("Ошибка парсинга широты: ", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка парсинга широты"})
				return
			}
			user.Longitude = longitude
		}
		avatarFile, _ := c.FormFile("avatar")
		if avatarFile != nil {
			user.Avatar = fmt.Sprintf("https://%s/media/%s%s", os.Getenv("SERVER_IP"), user.Name, strings.ToLower(filepath.Ext(avatarFile.Filename)))

			if err := utils.SaveAvatarFile(c, user.Name); err != nil {
				log.Println("Ошибка сохранения аватарки: ", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения аватарки"})
				return
			}

			req := internal.MediaServiceRequest{
				AvatarExt: filepath.Ext(avatarFile.Filename),
				UserName:  user.Name,
			}
			if err := addMediaRequest(req); err != nil {
				log.Println("Ошибка сохранения аватарки: ", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения аватарки"})
				return
			}

			if err := os.Remove(filepath.Join("/app", "glimpse", "media", fmt.Sprintf("%s%s", name, strings.ToLower(filepath.Ext(avatarFile.Filename))))); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения файла"})
				log.Println("Ошибка сохранения аватарки: ", err)
				return
			}
		}
		user.LastOnline = time.Now().UnixMilli()
		if err := utils.UserDb.Create(&user).Error; err != nil {
			log.Println("Ошибка сохранения данных в базу данных: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения данных в базу данных"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Вы успешно зарегестрированы!"})
	} else {
		c.JSON(http.StatusConflict, gin.H{"error": "Данное имя пользовтеля знаято введите другое"})
	}
}

func SignIn(c *gin.Context) {
	var request internal.AuthRequest
	var user internal.User
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Println("Ошибка обработки данных: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки данных"})
		return
	}

	if err := utils.UserDb.Where("name = ?", request.UserName).First(&user).Error; err != nil {
		log.Println("Пользователя с таким именем не существует: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Пользователя с таким именем не существует"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)) == nil {
		token, err := utils.GenerateJWTToken(user.Id, user.Name, time.UnixMilli(user.CreatedAt))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генрации токена"})
			log.Println("Ошибка генерации токена: ", err)
			return
		}
		c.JSON(http.StatusOK, token)
	} else {
		c.JSON(http.StatusConflict, gin.H{"error": "Вы ввели неверный пароль"})
	}
}

func DeleteUser(c *gin.Context) {
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
		log.Println("Ошибка получения пользователя: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения пользователя"})
		return
	}

	if strings.TrimSpace(user.Avatar) != "" {
		fileUrl, err := url.Parse(user.Avatar)
		if err != nil {
			log.Println("Ошибка прасинга url: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка прасинга url"})
			return
		}

		req := internal.MediaServiceRequest{
			AvatarExt: filepath.Ext(fileUrl.Path),
			UserName:  user.Name,
		}
		if err := deleteMediaRequest(req); err != nil {
			log.Println("Ошибка отправки запрос на удаления аватарки: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка отправки запрос на удаления аватарки"})
			return
		}
	}

	if err := utils.UserDb.Exec("DELETE FROM user_friends WHERE user_id = ? OR friend_id = ?", userId, userId).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления друзей"})
		log.Println("Ошибка удаления друзей: ", err)
		return
	}

	msgReq := internal.MsgServiceRequest{UserId: userId}
	msgs, err := deleteMessagesRequest(msgReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления сообщений"})
		return
	}
	mesgs := *msgs
	for _, e := range user.Friends {
		for i := range mesgs {
			e.SentMessages = append(e.SentMessages[:slices.Index(e.SentMessages, mesgs[i])], e.SentMessages[slices.Index(e.SentMessages, mesgs[i])+1:]...)
			e.ReceivedMessages = append(e.ReceivedMessages[:slices.Index(e.ReceivedMessages, mesgs[i])], e.ReceivedMessages[slices.Index(e.ReceivedMessages, mesgs[i])+1:]...)
		}
	}

	if err := utils.UserDb.Delete(&user).Error; err != nil {
		log.Println("Ошибка удаления учетной записи: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления учетной записи"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ваша учетная запись была успешно удалена"})
}

func UpdateUser(c *gin.Context) {
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
	if err := utils.UserDb.Preload("Friends.Friends").First(&user, userId).Error; err != nil {
		log.Println("Ошибка получения пользователя: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения пользоватей"})
		return
	}

	name := c.PostForm("name")
	if name != "" {
		user.Name = name
	}

	strPassword := c.PostForm("password")
	if strPassword != "" {
		password, err := bcrypt.GenerateFromPassword([]byte(strPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Println("Ошибка генерации пароля: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генрации пароля"})
			return
		}
		user.Password = string(password)
	}
	bio := c.PostForm("bio")
	if bio != "" {
		user.Bio = bio
	}
	strLatitude := c.PostForm("latitude")
	if strings.TrimSpace(strLatitude) != "" {
		latitude, err := strconv.ParseFloat(strLatitude, 64)
		if err != nil {
			log.Println("Ошибка парсинга широты: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка парсинга широты"})
			return
		}
		user.Latitude = latitude
	}
	strLongitude := c.PostForm("longitude")
	if strings.TrimSpace(strLongitude) != "" {
		longitude, err := strconv.ParseFloat(strLongitude, 64)
		if err != nil {
			log.Println("Ошибка парсинга широты: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка парсинга широты"})
			return
		}
		user.Longitude = longitude
	}
	avatarFile, _ := c.FormFile("avatar")
	if avatarFile != nil {
		if strings.TrimSpace(user.Avatar) != "" {
			user.Avatar = fmt.Sprintf("https://%s/media/%s%s", os.Getenv("SERVER_IP"), user.Name, strings.ToLower(filepath.Ext(avatarFile.Filename)))
			fileUrl, err := url.Parse(user.Avatar)
			if err != nil {
				log.Println("Ошибка прасинга url: ", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка прасинга url"})
				return
			}
			req := internal.MediaServiceRequest{
				AvatarExt: filepath.Ext(fileUrl.Path),
				UserName:  user.Name,
			}
			if err := deleteMediaRequest(req); err != nil {
				log.Println("Ошибка отправки запрос на удаления аватарки: ", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка отправки запрос на удаления аватарки"})
				return
			}

			if err := utils.SaveAvatarFile(c, user.Name); err != nil {
				log.Println("Ошибка сохранения аватарки: ", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения аватарки"})
				return
			}
			if err := addMediaRequest(req); err != nil {
				log.Println("Ошибка сохранения аватарки: ", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения аватарки"})
				return
			}

			if err := os.Remove(filepath.Join("/app", "glimpse", "media", fmt.Sprintf("%s%s", user.Name, strings.ToLower(filepath.Ext(avatarFile.Filename))))); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения файла"})
				log.Println("Ошибка сохранения аватарки: ", err)
				return
			}
		}
	}

	if d := c.PostForm("online"); d != "" {
		user.LastOnline = time.Now().UnixMilli()
	}

	if err := utils.UserDb.Save(&user).Error; err != nil {
		log.Println("Ошибка обновления учетной записи: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления учетной записи"})
		return
	}

	user.Id = userId

	c.JSON(http.StatusOK, user)
}

func GetUsers(c *gin.Context) {
	var (
		resps []internal.Users
		users []internal.User
	)
	if err := utils.UserDb.Find(&users).Error; err != nil {
		log.Println("Ошибка получения пользователей: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения пользователей"})
		return
	}
	for i := 0; len(users) > i; i++ {
		var resp internal.Users
		resp.Id = users[i].Id
		resp.Avatar = users[i].Avatar
		resp.Name = users[i].Name
		resp.Latitude = users[i].Latitude
		resp.Longitude = users[i].Longitude
		resps = append(resps, resp)
	}
	c.JSON(http.StatusOK, resps)
}

func WebSocket(c *gin.Context) {
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
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	cnn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Ошибка создания webSocket соединения: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания websocket соединения"})
		return
	}
	defer cnn.Close()

	var user internal.User

	if err := utils.UserDb.Preload("Friends.Friends").First(&user, userId).Error; err != nil {
		log.Println("Ошибка получения пользователя: ", err)
		cnn.WriteMessage(websocket.TextMessage, []byte("Ошибка получеения пользователя"))
		return
	}

	utils.TcpCns[user.Name] = cnn

	cnn.WriteMessage(websocket.TextMessage, []byte("WebSocket соединение успешно создано"))

	for {
		_, _, err := cnn.ReadMessage()
		if err != nil {
			log.Println("Ошибка получения сообщения: ", err)
			cnn.WriteMessage(websocket.TextMessage, []byte("Ошибка получения сообщения"))
			break
		}

		if err := utils.UserDb.Preload("Friends.Friends").First(&user, userId).Error; err != nil {
			log.Println("Ошибка получения пользовтеля: ", err)
			cnn.WriteMessage(websocket.TextMessage, []byte("Ошибка получеения пользователя"))
			return
		}

		if data, err := json.Marshal(&user); err != nil {
			log.Println("Ошибка преобразования типов: ", err)
			cnn.WriteMessage(websocket.TextMessage, []byte("Ошибка преобразования типов"))
			break
		} else {
			for _, o := range user.Friends {
				for k, v := range utils.TcpCns {
					if o.Name == k {
						var friend internal.User
						if err := utils.UserDb.Where("name = ?", k).Preload("Friends.Friends").First(&friend).Error; err != nil {
							log.Println("Ошибка получения пользовтеля: ", err)
							cnn.WriteMessage(websocket.TextMessage, []byte("Ошибка получеения пользователя"))
							return
						}
						friendData, err := json.Marshal(friend)
						if err != nil {
							log.Println("Ошибка кодирования json: ", err)
							cnn.WriteMessage(websocket.TextMessage, []byte("Ошибка кодирования json"))
							return
						}
						v.WriteMessage(websocket.TextMessage, friendData)
					}
				}
			}
			if err := cnn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Println("Ошибка отправки json: ", err)
				cnn.WriteMessage(websocket.TextMessage, []byte("Ошибка отпрваки json"))
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "WebSocket соединение успешно разорвано!"})
}
