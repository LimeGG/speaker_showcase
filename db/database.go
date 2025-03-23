package db

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

type DB struct {
	Db *gorm.DB
}

var DataBase DB

// Migrate применяет миграции для базы данных
func Migrate(DB *gorm.DB) {
	err := DB.AutoMigrate(
		&User{},
		&Speaker{},
		&PersonalAccount{},
		&Course{},
		&ResetPassword{},
		&PasswordReset{},
		&ForgotPasswordRequest{},
	)
	if err != nil {
		log.Fatal(err)
	}

}

// InitDb инициализирует базу данных и сохраняет ее в глобальной переменной
func InitDb() *gorm.DB {
	//dsn := "host=localhost user=postgres password=08092003a dbname=PostgreSQL port=5432 "
	dsn := "host=localhost user=postgres password=08092003 dbname=speaker port=5432 sslmode=disable TimeZone=Europe/Moscow"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	Migrate(db)
	fmt.Println("database initialized")

	// Сохраняем инициализированную базу данных в глобальной переменной DataBase
	DataBase.Db = db

	return db
}
