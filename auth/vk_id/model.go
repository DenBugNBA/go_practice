package main

// AuthConfig Конфиг для авторизации через VK
type AuthConfig struct {
	// ClientID идентификатор приложения
	ClientID string
	// AuthURL URL для авторизации
	AuthURL string
	// TokenURL URL для получения токенов
	TokenURL string
	// UserInfoURL URL для получения данных пользователя
	UserInfoURL string
	// RedirectURL URL для перенаправления пользователей после авторизации
	RedirectURL string
}

type AuthResp struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	UserId       int    `json:"user_id"`
	State        string `json:"state"`
	Scope        string `json:"scope"`

	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type UserInfo struct {
	User             *User  `json:"user"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type User struct {
	UserId    string `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Avatar    string `json:"avatar"`
	Email     string `json:"email"`
	Sex       int    `json:"sex"`
	Verified  bool   `json:"verified"`
	Birthday  string `json:"birthday"`
}
