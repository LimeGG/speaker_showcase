package utils

import (
	"bytes"
	. "cms/db"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var mySigningKey = []byte("hash123132-data34141")

func CreateJWTToken(userID uint, email string) (string, error) {
	// Определяем время жизни токена
	expirationTime := time.Now().Add(24 * time.Hour)

	// Создаем клаймы для токена
	claims := &JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	secretKey := mySigningKey

	// Создаем токен с алгоритмом подписи
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func VerifyToken(tokenStr string) (*JWTClaims, error) {
	claims := &JWTClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неверный метод подписи: %v", token.Header["alg"])
		}
		return mySigningKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("неверный токен")
	}

	return claims, nil
}

func CheckPhone(phone string) (string, bool) {
	res := strings.ReplaceAll(phone, " ", "")
	res = strings.ReplaceAll(res, "+", "")

	ph, err := strconv.ParseInt(res, 10, 64)
	if err != nil || len(res) < 11 {
		fmt.Println(err)
		return "", false
	}

	return strconv.Itoa(int(ph)), true
}

func CheckEmail(email string) (string, bool) {
	strs := strings.Split(email, "@")
	if len(strs) > 1 {
		if len(strs[0]) > 0 && len(strs[1]) > 2 {
			for i := range strs[1] {
				if string(strs[1][i]) == "." {
					return email, true
				}
			}
		}
	}
	return "", false
}

func HashPassword(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	return hex.EncodeToString(hash.Sum(nil))
}

func GenerateResetToken() (string, error) {
	bytes := make([]byte, 16) // 16 байт = 32 символа в hex
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

type EmailRequest struct {
	Action   string `json:"action"`
	Letter   Letter `json:"letter"`
	Group    string `json:"group"`
	Email    string `json:"email"`
	SendWhen string `json:"sendwhen"`
	APIKey   string `json:"apikey"`
}

type Letter struct {
	Message   Message `json:"message"`
	Subject   string  `json:"subject"`
	FromEmail string  `json:"from.email"`
}

type Message struct {
	HTML string `json:"html"`
}

func SendResetEmail(email, resetURL string) error {
	// Создаем HTML-версию письма
	htmlContent := fmt.Sprintf(`<html><body><p>Для сброса пароля перейдите по ссылке: <a href="%s">%s</a></p></body></html>`, resetURL)

	apiKey := "19mb7Ghr6T7beNlVXc5w5vKgI9u9ezE2slEhJwXalCcuL3df8n7qO4NzQA6zgqZL2gBsP7exr"

	// Создаем экземпляр структуры с данными для отправки
	emailRequest := EmailRequest{
		Action: "issue.send",
		Letter: Letter{
			Message: Message{
				HTML: htmlContent,
			},
			Subject:   "Сброс пароля",
			FromEmail: "miruta.a@gs.donstu.ru",
		},
		Group:    "personal",
		Email:    email,
		SendWhen: "now",
		APIKey:   apiKey,
	}

	// Преобразуем данные в JSON
	jsonData, err := json.Marshal(emailRequest)
	if err != nil {
		return fmt.Errorf("ошибка при сериализации данных: %v", err)
	}

	url := "https://api.sendsay.ru/general/api/v100/json/x_1732522802640968"

	// Создаем новый POST-запрос
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("ошибка при создании запроса: %v", err)
	}

	// Устанавливаем заголовки запроса
	req.Header.Set("Content-Type", "application/json")

	// Создаем HTTP-клиент с таймаутом
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Отправляем запрос
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка при отправке запроса: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ошибка при отправке письма: статус %d", resp.StatusCode)
	}

	fmt.Printf("Письмо на %s с URL: %s успешно отправлено\n", email, resetURL)
	return nil
}
