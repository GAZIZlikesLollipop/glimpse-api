package internal

type Message struct {
	Id         int64   `json:"id" gorm:"primaryKey"`
	Content    string  `json:"content"`
	SenderId   *uint64 `json:"senderId"`
	ReceiverId *uint64 `json:"receiverId"`
	CreatedAt  int64   `json:"created_at" gorm:"column:created_at;type:bigint;autoCreateTime:milli"`
	UpdatedAt  int64   `json:"updated_at" gorm:"column:updated_at;type:bigint;autoUpdateTime:milli"`
}

type User struct {
	Id               int64     `json:"id" gorm:"primaryKey"`
	Name             string    `json:"name"`
	Password         string    `json:"password"`
	Avatar           string    `json:"avatar"`
	Bio              string    `json:"bio"`
	Latitude         float64   `json:"latitude"`
	Longitude        float64   `json:"longitude"`
	LastOnline       int64     `json:"lastOnline"`
	Friends          []User    `json:"friends" gorm:"many2many:user_friends;"`
	SentMessages     []Message `json:"sentMessages" gorm:"foreignKey:SenderId;constraint:OnDelete:CASCADE"`
	ReceivedMessages []Message `json:"receivedMessages" gorm:"foreignKey:ReceiverId;constraint:OnDelte:CASCADE"`
	CreatedAt        int64     `json:"created_at" gorm:"column:created_at;type:bigint;autoCreateTime:milli"`
	UpdatedAt        int64     `json:"updated_at" gorm:"column:updated_at;type:bigint;autoUpdateTime:milli"`
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
