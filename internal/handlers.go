package internal

import (
	"api/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func SignUp(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Println("Ошибка обработки тела запроса: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки тела запроса"})
		return
	}

	var count int64
	if err := utils.Db.Model(&User{}).Where("name = ?", user.Name).Count(&count).Error; err == nil {
		log.Println("Ошибка получения пользователя: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения пользователя"})
		return
	}

	if count < 1 {
		password, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Println("Ошибка генерации пароля: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генрации пароля"})
			return
		}
		user.Password = string(password)
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
	var request AuthRequest
	var user User
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Println("Ошибка обработки данных: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки данных"})
		return
	}

	if err := utils.Db.Where("name = ?", request.UserName).First(&user).Error; err != nil {
		log.Println("Ошибка получения пользовтеля: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения пользовтеля"})
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
		c.JSON(http.StatusConflict, gin.H{"error": "Пользователя с таким именем не существует"})
	}
}

func AuthMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header != "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неправильный заголовок"})
			c.Abort()
			return
		}
		if len(header) < 7 || header[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неправильный заголовок"})
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
