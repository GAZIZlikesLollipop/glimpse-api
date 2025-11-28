package internal

import "github.com/lib/pq"

type Message struct {
	Id         int64  `json:"id" gorm:"primaryKey"`
	Content    string `json:"content"`
	SenderId   *int64 `json:"senderId"`
	ReceiverId *int64 `json:"receiverId"`
	CreatedAt  int64  `json:"created_at" gorm:"column:created_at;type:bigint;autoCreateTime:milli"`
	UpdatedAt  int64  `json:"updated_at" gorm:"column:updated_at;type:bigint;autoUpdateTime:milli"`
}

type User struct {
	Id               int64         `json:"id" gorm:"primaryKey"`
	Name             string        `json:"name"`
	Password         string        `json:"password"`
	Avatar           string        `json:"avatar"`
	Bio              string        `json:"bio"`
	Latitude         float64       `json:"latitude"`
	Longitude        float64       `json:"longitude"`
	LastOnline       int64         `json:"lastOnline"`
	Friends          []User        `json:"friends" gorm:"many2many:user_friends;"`
	SentMessages     pq.Int64Array `json:"sentMessages" gorm:"type:bigint[]"`
	ReceivedMessages pq.Int64Array `json:"receivedMessages" gorm:"type:bigint[]"`
	CreatedAt        int64         `json:"created_at" gorm:"column:created_at;type:bigint;autoCreateTime:milli"`
	UpdatedAt        int64         `json:"updated_at" gorm:"column:updated_at;type:bigint;autoUpdateTime:milli"`
}

type AuthRequest struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

type Users struct {
	Id        int64   `json:"id"`
	Name      string  `json:"name"`
	Avatar    string  `json:"avatar"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type MsgServiceRequest struct {
	UserId   int64  `json:"user_id"`
	FriendId *int64 `json:"friend_id"`
}

type UserSeriveRequest struct {
	SenderId   int64 `json:"sender_id"`
	ReceiverId int64 `json:"receiver_id"`
	MsgId      int64 `json:"user_msg_id"`
}

type MediaServiceRequest struct {
	AvatarExt string `json:"avatar_ext"`
	UserName  string `json:"user_name"`
}
