package internal

import (
	"time"
)

type Message struct {
	Id         int64     `json:"id" gorm:"primaryKey"`
	Content    string    `json:"content"`
	IsChecked  bool      `json:"isChecked"`
	SenderId   *uint64   `json:"senderId"`
	ReceiverId *uint64   `json:"receiverId"`
	CreatedAt  time.Time `json:"crated_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type User struct {
	Id         int64     `json:"id" gorm:"primaryKey"`
	Name       string    `json:"name"`
	Password   string    `json:"password"`
	Avatar     string    `json:"avatar"`
	Bio        string    `json:"bio"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	LastOnline time.Time `json:"lastOnline"`
	Friends    []User    `json:"friends" gorm:"many2many:user_friends;"`
	// SentMessages     []Message `json:"sentMessages" gorm:"foreignKey:SenderId"`
	// ReceivedMessages []Message `json:"receivedMessages" gorm:"foreignKey:ReceiverId"`
	CreatedAt time.Time `json:"crated_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AuthRequest struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}
