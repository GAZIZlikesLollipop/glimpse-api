package internal

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Message struct {
	Id        int64     `json:"id" gorm:"primaryKey"`
	ChatID    int64     `json:"chat_id" gorm:"not null"`
	Content   string    `json:"content"`
	Owner     User      `json:"owner"`
	CreatedAt time.Time `json:"crated_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FriendData struct {
	Id        int64        `json:"id"`
	Name      string       `json:"name"`
	Avatar    string       `json:"avatar"`
	City      string       `json:"city"`
	Friends   []FriendData `json:"friends"`
	CreatedAt time.Time    `json:"crated_at"`
}

type FirendUser struct {
	Name       string       `json:"name"`
	Avatar     string       `json:"avatar"`
	Latitude   float64      `json:"latitude"`
	Longitude  float64      `json:"longitude"`
	LastOnline time.Time    `json:"lastOnline"`
	Friends    []FriendData `json:"friends"`
	CreatedAt  time.Time    `json:"crated_at"`
}

type Friend struct {
	Id       int64     `json:"id" gorm:"primaryKey"`
	Data     User      `json:"user"`
	Messages []Message `json:"messages" gorm:"foreignKey:FriendID"`
}

type User struct {
	Id         int64     `json:"id" gorm:"primaryKey"`
	Name       string    `json:"name"`
	Avatar     string    `json:"avatar"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	LastOnline time.Time `json:"lastOnline"`
	Friends    []Friend  `json:"friends" gorm:"foreignKey:UserID"`
	CreatedAt  time.Time `json:"crated_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Claims struct {
	UserId    int64     `json:"user_id"`
	UserName  string    `json:"userName"`
	CreatedAt time.Time `json:"crated_at"`
	jwt.RegisteredClaims
}
