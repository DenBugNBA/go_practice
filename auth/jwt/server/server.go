package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("my_secret_key")

type Credentials struct {
	Username string `json:"username"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/login", LoginHandler)
	r.With(JWTMiddleware).Get("/protected", ProtectedHandler)

	fmt.Println("starting server on 8080")
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}

// LoginHandler POST /login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil || creds.Username == "" {
		http.Error(w, "Неверные данные", http.StatusBadRequest)
		return
	}

	// Генерация токена
	expirationTime := time.Now().Add(15 * time.Minute)
	claims := &Claims{
		Username: creds.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    "myapp",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Ошибка генерации токена", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

// ProtectedHandler GET /protected
func ProtectedHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)
	w.Write([]byte("Привет, " + username + "! Это защищённый ресурс.\n"))
}

// JWTMiddleware Middleware для проверки JWT
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Токен не предоставлен", http.StatusUnauthorized)
			return
		}

		tokenStr := authHeader[len("Bearer "):]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Неверный токен", http.StatusUnauthorized)
			return
		}

		// Передаём имя пользователя дальше через контекст
		ctx := r.Context()
		ctx = contextWithUsername(ctx, claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Упрощённый способ передать username в context
func contextWithUsername(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, "username", username)
}
