package db

import (
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"time"
)

type User struct {
	gorm.Model
	Name            string    `json:"name" gorm:"size:50"`                // Ограничение в 50 символов
	Surname         string    `json:"surname" gorm:"size:50"`             // Ограничение в 50 символов
	LastName        string    `json:"last_name" gorm:"size:50"`           // Ограничение в 50 символов
	PhoneNumber     string    `json:"phone_number" gorm:"size:15;unique"` // Ограничение в 15 символов и уникальность
	Email           string    `json:"mail" gorm:"size:100;unique"`        // Ограничение в 100 символов и уникальность
	Password        string    `json:"-" gorm:"size:255"`                  // Ограничение в 255 символов
	Age             uint8     `json:"age"`
	AvatarUrl       string    `json:"avatar" gorm:"size:255"` // Ограничение в 255 символов
	Sex             bool      `json:"sex"`
	Role            string    `json:"role"` // Ограничение в 20 символов
	SubscribeStatus bool      `json:"subscribe_status"`
	SubscribeStart  time.Time `json:"subscribe_start"`
	SubscribeEnd    time.Time `json:"subscribe_end"`
	IsSchool        bool      `json:"school"`
	Speaker         Speaker   `gorm:"foreignKey:UserID"`
}

type Speaker struct {
	gorm.Model
	UserID          uint            `json:"user_id" gorm:"unique"` // Внешний ключ для связи с User
	IsSubscribed    bool            `json:"is_subscribed"`
	PersonalAccount PersonalAccount `gorm:"foreignKey:SpeakerID"` // Связь один к одному
	Courses         []Course        `gorm:"foreignKey:SpeakerID"` // Связь один ко многим
}

type PersonalAccount struct {
	gorm.Model
	SpeakerID   uint           `json:"speaker_id" gorm:"unique"` // Внешний ключ для связи с Speaker
	Skills      datatypes.JSON `json:"skills"`
	Description string         `gorm:"size:255"`
	WorkPoint   datatypes.JSON `json:"work_point"`
	Contacts    datatypes.JSON `gorm:"type:jsonb" json:"contacts"`
}

type Course struct {
	gorm.Model
	SpeakerID   uint           `json:"speaker_id" gorm:"unique"`
	Link        string         `json:"link" gorm:"size:255"`        // Ограничение в 255 символов
	Name        string         `json:"name" gorm:"size:100"`        // Ограничение в 100 символов
	Description string         `json:"description" gorm:"size:500"` // Ограничение в 500 символов
	Program     datatypes.JSON `json:"program"`
	Tools       datatypes.JSON `json:"tools"`
}

type JWTClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type ForgotPasswordRequest struct {
	Mail string `json:"mail"`
}

type ResetPassword struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type PasswordReset struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null"`
	Token     string    `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time
}
