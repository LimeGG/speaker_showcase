package auth

import (
	. "cms/db"
	. "cms/utils"
	"fmt"
	"github.com/labstack/echo"
	"gorm.io/gorm"
	"log"
	"net/http"
	"time"
)

func InitAuth(auth *echo.Group) {
	auth.GET("/getuser", GetUser)
	auth.POST("/register", Register)
	auth.POST("/login", Login)
	auth.PUT("/lkuserupdate", UpdateUser)
	auth.POST("/forgotpassword", ForgotPassword)
	auth.POST("/resetpassword/:token", ResetsPassword)
}

func GetUser(c echo.Context) error {

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

	var user []User

	if err := DataBase.Db.Where("id = ?", JWTClaim.UserID).Find(&user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, user)
}

func Register(c echo.Context) error {
	data := new(User)

	// Привязка данных пользователя
	if err := c.Bind(data); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data"})
	}

	// Проверка обязательных полей
	if data.Email == "" || data.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email and password are required"})
	}

	// Проверка на существование пользователя с таким же email
	var existingUser User
	if err := DataBase.Db.Where("email = ?", data.Email).First(&existingUser).Error; err == nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": "User already exists"})
	}

	// Хеширование пароля
	data.Password = HashPassword(data.Password)

	// Сохраняем пользователя в базе данных
	if err := DataBase.Db.Create(&data).Error; err != nil {
		log.Printf("Failed to insert user: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}

	// Генерация JWT токена с новым ID пользователя
	jwt, err := CreateJWTToken(data.ID, data.Email)
	if err != nil {
		log.Fatalf("Error creating JWT token: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create JWT token"})
	}

	// Возвращаем JWT токен клиенту
	return c.JSON(http.StatusOK, map[string]string{"jwt": jwt})
}

func Login(c echo.Context) error {

	type LoginRequest struct {
		Mail     string `json:"mail" form:"mail"`
		Password string `json:"password" form:"password"`
	}

	var req LoginRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data"})
	}

	if req.Mail == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email and password are required"})
	}

	var user User
	if err := DataBase.Db.Where("mail = ?", req.Mail).First(&user).Error; err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid email or password"})
	}

	if user.Password != HashPassword(req.Password) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid password"})
	}

	jwt, err := CreateJWTToken(user.ID, user.Email)
	if err != nil {
		log.Printf("Failed to create JWT token: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}

	return c.JSON(http.StatusOK, map[string]string{"jwt": jwt})

}

func UpdateUser(c echo.Context) error {

	jwt := c.Request().Header.Get("Authorization")
	if jwt == "" {
		return c.JSON(http.StatusUnauthorized, "No jwt header")
	}
	JWTClaim, err := VerifyToken(jwt)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, err)
	}

	// Получение данных для обновления
	var updateData User
	if err := c.Bind(&updateData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data"})
	}

	var existingUser User
	if err := DataBase.Db.Where("id = ?", JWTClaim.UserID).First(&existingUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
	}

	// Обновляем существующую анкету
	if err := DataBase.Db.Model(&existingUser).Where("id = ?", JWTClaim.UserID).Updates(updateData).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update User"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "User updated successfully"})

}

//func ForgotPassword(c echo.Context) error {
//
//	var request ResetPassword
//	if err := c.Bind(&request); err != nil {
//		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data"})
//	}
//
//	if err := c.Validate(&request); err != nil {
//		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid token or password"})
//	}
//
//	var resetEntry PasswordReset
//	if err := DataBase.Db.Where("token = ?", request.Token).First(&resetEntry).Error; err != nil {
//		if err == gorm.ErrRecordNotFound {
//			return c.JSON(http.StatusNotFound, map[string]string{"error": "Invalid or expired token"})
//		}
//		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
//	}
//
//}

func ForgotPassword(c echo.Context) error {

	var request ForgotPasswordRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data"})
	}

	var user User
	if err := DataBase.Db.Where("mail = ?", request.Mail).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
	}

	// Генерация токена для восстановления пароля
	token, err := GenerateResetToken()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate reset token"})
	}

	// Сохранение токена в базе данных (можно использовать отдельную таблицу)
	resetEntry := PasswordReset{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour), // Токен будет действителен 1 час
	}
	if err := DataBase.Db.Create(&resetEntry).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save reset token"})
	}

	// Отправка email с инструкциями
	resetURL := fmt.Sprintf("http://localhost:6000/api/v1/auth/resetpassword/%s", token)
	if err := SendResetEmail(user.Email, resetURL); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to send reset email"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Password reset email sent"})
}

func ResetsPassword(c echo.Context) error {
	token := c.Param("token")

	type ResetPasswordRequest struct {
		Token    string `json:"token" validate:"required"`
		Password string `json:"password" validate:"required"`
	}

	var request ResetPasswordRequest
	request.Token = token
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data"})
	}

	//if err := c.Validate(&request); err != nil {
	//	return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid token or password"})
	//}

	// Проверяем токен
	var resetEntry PasswordReset
	if err := DataBase.Db.Where("token = ?", request.Token).First(&resetEntry).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Invalid or expired token"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
	}

	if resetEntry.ExpiresAt.Before(time.Now()) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Token has expired"})
	}

	// Обновляем пароль пользователя
	hashedPassword := HashPassword(request.Password)
	if err := DataBase.Db.Model(&User{}).Where("id = ?", resetEntry.UserID).Update("password", hashedPassword).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update password"})
	}

	// Удаляем токен после успешного сброса
	DataBase.Db.Delete(&resetEntry)

	return c.JSON(http.StatusOK, map[string]string{"message": "Password reset successfully"})
}
