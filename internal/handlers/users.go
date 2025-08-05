package handlers

import (
	"api/internal"
	"api/utils"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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

	if err := utils.Db.Preload("SentMessages").Preload("ReceivedMessages").Preload("Friends").First(&user, userId).Error; err != nil {
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
	if err := utils.Db.Model(&internal.User{}).Where("name = ?", user.Name).Count(&count).Error; err != nil {
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
			var err error
			user.Avatar, err = utils.SaveAvatarFile(c, user.Name)
			if err != nil {
				log.Println("Ошибка сохранения файла: ", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения файла"})
				return
			}
		}
		user.LastOnline = time.Now()
		if err := utils.Db.Create(&user).Error; err != nil {
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

	if err := utils.Db.Where("name = ?", request.UserName).First(&user).Error; err != nil {
		log.Println("Пользователя с таким именем не существует: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Пользователя с таким именем не существует"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)) == nil {
		token, err := utils.GenerateJWTToken(user.Id, user.Name, user.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генрации токена"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": "Вы успешно вошли в свою учетную запись", "token": token})
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

	if err := utils.Db.Preload("SentMessages").Preload("ReceivedMessages").Preload("Friends").First(&user, userId).Error; err != nil {
		log.Println("Ошибка получения пользователя: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения пользователя"})
		return
	}

	if strings.TrimSpace(user.Avatar) != "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Println("Ошибка поулчения домашней диреткории: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения пути к домшней директории"})
			return
		}
		fileUrl, err := url.Parse(user.Avatar)
		if err != nil {
			log.Println("Ошибка прасинга url: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка прасинга url"})
			return
		}
		if err := os.Remove(filepath.Join(homeDir, fileUrl.Path)); err != nil {
			log.Println("Ошибка удаления файла: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления файла"})
			return
		}
	}

	if err := utils.Db.Model(&user).Association("Friends").Clear(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления друзей"})
		log.Println("Ошибка удаления друзей: ", err)
		return
	}

	if err := utils.Db.Model(&user).Association("SentMessages").Clear(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления сообщений"})
		log.Println("Ошибка удаления сообщений: ", err)
		return
	}

	if err := utils.Db.Model(&user).Association("ReceivedMessages").Clear(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления сообщений"})
		log.Println("Ошибка удаления сообщений: ", err)
		return
	}

	if err := utils.Db.Delete(&user).Error; err != nil {
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
	if err := utils.Db.First(&user, userId).Error; err != nil {
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
		var err error
		if user.Avatar != "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				log.Fatalln("Ошибка поулчения домашней диреткории: ", err)
			}
			urlPath, err := url.Parse(user.Avatar)
			if err != nil {
				log.Println("Ошибка полчения урл пути: ", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения урл пути"})
				return
			}
			if err := os.Remove(filepath.Join(homeDir, urlPath.Path)); err != nil {
				log.Println("Ошибка удаления файла: ", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления файла"})
				return
			}
		}
		user.Avatar, err = utils.SaveAvatarFile(c, user.Name)
		if err != nil {
			log.Println("Ошибка сохранения файла: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения файла"})
			return
		}
	}

	user.LastOnline = time.Now()

	if err := utils.Db.Save(&user).Error; err != nil {
		log.Println("Ошибка обновления учетной записи: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления учетной записи"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Успешное обновление учетной записи"})
}

func AuthMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный заголовок авторизации"})
			c.Abort()
			return
		}
		if len(header) < 7 || header[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный формат токена"})
			c.Abort()
			return
		}
		token := header[7:]
		claims, err := utils.ValidateJWTToken(token)
		if err != nil {
			log.Println("Недействительный токен jwt: ", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Недействительный токен jwt"})
			c.Abort()
			return
		}
		c.Set("userId", claims.UserId)
		c.Set("userName", claims.UserName)
		c.Next()
	}
}
