package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("my_secret_key") // Секрет для подписи

// Claims Пользовательская структура для payload
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func main() {
	// Создание токена
	tokenString, err := GenerateJWT("johndoe")
	if err != nil {
		panic(err)
	}
	fmt.Println("JWT:", tokenString)

	// Проверка токена
	username, err := ValidateJWT(tokenString)
	if err != nil {
		fmt.Println("Ошибка валидации:", err)
	} else {
		fmt.Println("Пользователь:", username)
	}
}

// GenerateJWT Генерация токена
func GenerateJWT(username string) (string, error) {
	expirationTime := time.Now().Add(15 * time.Minute)
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    "example.com",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// ValidateJWT Валидация токена
func ValidateJWT(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return "", err
	}
	if !token.Valid {
		return "", fmt.Errorf("недопустимый токен")
	}
	return claims.Username, nil
}
