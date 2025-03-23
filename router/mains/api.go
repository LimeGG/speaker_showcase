package mains

import (
	. "cms/db"
	. "cms/utils"
	"errors"
	"github.com/labstack/echo"
	"gorm.io/gorm"
	"net/http"
)

func InitApi(api *echo.Group) {
	api.GET("/getanketa", GetAncl)
	api.POST("/createanketa", CreateAncl)
	api.PUT("/updateanketa", UpdateAncl)
}

func GetAncl(c echo.Context) error {
	// Извлекаем JWT из заголовка
	jwt := c.Request().Header.Get("Authorization")
	if jwt == "" {
		return c.JSON(http.StatusUnauthorized, "No jwt header")
	}

	// Проверяем токен и извлекаем данные пользователя
	JWTClaim, err := VerifyToken(jwt)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, err.Error())
	}

	println(JWTClaim.UserID)

	var anketa []PersonalAccount

	// Находим записи анкеты для текущего пользователя
	if err := DataBase.Db.Where("user_id = ?", JWTClaim.UserID).Find(&anketa).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	// Проверяем, есть ли анкеты
	if len(anketa) == 0 {
		return c.JSON(http.StatusOK, "Анкеты нет")
	}

	return c.JSON(http.StatusOK, anketa)
}

func CreateAncl(c echo.Context) error {
	jwt := c.Request().Header.Get("Authorization")
	if jwt == "" {
		return c.JSON(http.StatusUnauthorized, "No JWT header")
	}

	JWTClaim, err := VerifyToken(jwt)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, err)
	}

	println("Extracted UserID:", JWTClaim.UserID)
	if JWTClaim.UserID == 0 {
		return c.JSON(http.StatusUnauthorized, "Invalid JWT token: UserID is 0")
	}

	// Проверка, существует ли пользователь
	var user User
	if err := DataBase.Db.First(&user, JWTClaim.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "User not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
	}

	// Привязка данных анкеты
	data := new(PersonalAccount)
	if err := c.Bind(data); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data"})
	}

	// Устанавливаем UserID из JWT
	data.UserID = JWTClaim.UserID

	if err := DataBase.Db.Create(&data).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}

	return c.JSON(http.StatusCreated, data)
}

func UpdateAncl(c echo.Context) error {
	jwt := c.Request().Header.Get("Authorization")
	if jwt == "" {
		return c.JSON(http.StatusUnauthorized, "No jwt header")
	}
	JWTClaim, err := VerifyToken(jwt)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, err)
	}

	// Получение данных для обновления
	var updateData Ancl
	if err := c.Bind(&updateData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data"})
	}

	// Проверяем существующую анкету
	var existingAncl Ancl
	if err := DataBase.Db.Where("user_id = ?", JWTClaim.UserID).First(&existingAncl).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Ancl not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
	}

	// Обновляем существующую анкету
	if err := DataBase.Db.Model(&existingAncl).Where("user_id = ?", JWTClaim.UserID).Updates(updateData).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update ancl"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Ancl updated successfully"})
}
