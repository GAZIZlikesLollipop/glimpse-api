package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"log"
	"os"
	"strconv"
	"time"

	"api/internal"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

var Db *gorm.DB

func GenerateJWTPrivateKey() error {
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

func GenerateJWT(userId int64, userName string, createdAt time.Time) (string, error) {
	secretKey, _ := base64.URLEncoding.DecodeString(os.Getenv("JWT_SCRET_KEY"))
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &internal.Claims{
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
