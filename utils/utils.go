package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var Db *gorm.DB

type Claims struct {
	UserId    int64     `json:"user_id"`
	UserName  string    `json:"userName"`
	CreatedAt time.Time `json:"crated_at"`
	jwt.RegisteredClaims
}

func GenerateJWTSecretKey() error {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Println("Ошибка генрации приватного ключа: ", err)
		return err
	}

	keyDER, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		log.Println("Ошибка преобраовния private key: ", err)
		return err
	}

	keyPM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyDER,
	})

	if err := os.Setenv("JWT_SECRET_KEY", string(keyPM)); err != nil {
		log.Println("Ошибка сохранения сохренния привтного клюа в премнную: ", err)
		return err
	}

	return nil
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
	secretKey := os.Getenv("JWT_SECRET_KEY")
	var claims Claims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", t.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("Введенные токен не валидный")
	}

	return &claims, nil
}

func SaveAvatarFile(
	c *gin.Context,
	name string,
) (string, error) {
	var filePath string
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Println("Ошибка поулчения домашней диреткории: ", err)
		return "", err
	}
	absolutePath := filepath.Join(homeDir, "glimpse-avatars")
	if err := os.MkdirAll(absolutePath, 0755); err != nil {
		log.Println("Ошибка создания диреткории")
		return "", err
	}
	file, err := c.FormFile("avatar")
	if err != nil {
		log.Println("Ошибка обработки файла: ", err)
		return "", err
	}

	filePath = fmt.Sprintf("%s-%s%s", name, uuid.New(), strings.ToLower(filepath.Ext(file.Filename)))

	if err := c.SaveUploadedFile(file, filepath.Join(absolutePath, filePath)); err != nil {
		log.Println("Ошибка сохранения файла: ", err)
		return "", err
	}

	var addr string

	ipaces, err := net.InterfaceAddrs()
	if err != nil {
		log.Println("Ошибка поулчения интерфейсов: ", err)
		return "", err
	}

	for i := 0; len(ipaces) > i; i++ {
		if ipaces[i].String()[0:6] == "192.168" {
			for ind := 0; len(ipaces[i].String()) > ind; ind++ {
				if string(ipaces[i].String()[ind]) != "/" {
					addr += string(ipaces[i].String()[ind])
				} else {
					break
				}
			}
			break
		} else {
			continue
		}
	}

	return fmt.Sprintf("https://%s:8080/avatars/%s", addr, filePath), nil
}
