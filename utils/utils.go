package utils

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var UserDb *gorm.DB
var MsgDb *gorm.DB

var TcpCns map[string]*websocket.Conn

func SaveAvatarFile(
	c *gin.Context,
	name string,
) error {
	absolutePath := filepath.Join("/app", "glimpse", "media")
	if err := os.MkdirAll(absolutePath, 0755); err != nil {
		log.Println("Ошибка создания диреткории: ", err)
		return err
	}
	file, err := c.FormFile("avatar")
	if err != nil {
		log.Println("Ошибка обработки файла: ", err)
		return err
	}

	fileName := fmt.Sprintf("%s%s", name, strings.ToLower(filepath.Ext(file.Filename)))

	if err := c.SaveUploadedFile(file, filepath.Join(absolutePath, fileName)); err != nil {
		log.Println("Ошибка сохранения файла: ", err)
		return err
	}

	return nil
}

type Claims struct {
	UserId    int64     `json:"user_id"`
	UserName  string    `json:"userName"`
	CreatedAt time.Time `json:"crated_at"`
	jwt.RegisteredClaims
}

func DecodeJWTSecretKey() (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(os.Getenv("JWT_SECRET_KEY")))

	if block == nil || block.Type != "EC PRIVATE KEY" {
		return nil, fmt.Errorf("не удалось декодировать PEM-блок приватного ключа или неверный тип")
	}

	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		log.Println("Ошибка прасинга pem строки: ", err)
		return nil, err
	}

	return privateKey, nil
}

func GenerateJWTToken(userId int64, userName string, createdAt time.Time) (string, error) {
	secretKey, err := DecodeJWTSecretKey()
	if err != nil {
		return "", err
	}
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserId:    userId,
		UserName:  userName,
		CreatedAt: createdAt,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "your-gin-crud-app",
			Subject:   strconv.FormatInt(userId, 10), // Преобразуем int в string для Subject
			Audience:  []string{"users"},
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenString, err := t.SignedString(secretKey)
	if err != nil {
		log.Println("Ошибка преобрзаовния токена в string: ", err)
		return "", err
	}

	return tokenString, nil
}

func ValidateJWTToken(tokenString string) (*Claims, error) {
	var claims Claims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", t.Header["alg"])
		}
		token, err := DecodeJWTSecretKey()
		if err != nil {
			log.Println("Ошибка декодирования токена: ", err)
			return "", nil
		}
		return &token.PublicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("введенные токен не валидный")
	}

	return &claims, nil
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
		claims, err := ValidateJWTToken(token)
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
