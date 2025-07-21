package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
