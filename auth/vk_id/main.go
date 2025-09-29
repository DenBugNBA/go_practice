package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	// подтягивает переменные окружения из .env
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog/log"
)

var (
	vkAuthConfig = &AuthConfig{
		ClientID:    os.Getenv("VK_CLIENT_ID"),
		AuthURL:     os.Getenv("VK_AUTH_URL"),
		TokenURL:    os.Getenv("VK_TOKEN_URL"),
		UserInfoURL: os.Getenv("VK_USER_INFO_URL"),
		RedirectURL: os.Getenv("VK_REDIRECT_URI"),
	}
)

func main() {
	port := os.Getenv("SERVER_PORT")

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", indexHandler)
	r.Get("/login", loginHandler)
	r.Get("/auth", authHandler)

	log.Info().Msg(fmt.Sprintf("server is running at port: %s", port))
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), r)
	if err != nil {
		log.Err(err).Msg("server stopped with err")
	}
}

func indexHandler(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte(`<a href="/login">Login with VK</a>`))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	authURL, err := buildAuthURL()
	if err != nil {
		panic(err)
	}
	http.Redirect(w, r, authURL, http.StatusFound)
}

func buildAuthURL() (string, error) {
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", vkAuthConfig.ClientID)
	params.Set("code_challenge", generateCodeChallenge(generateCodeVerifier()))
	params.Set("code_challenge_method", "S256")
	params.Set("redirect_uri", vkAuthConfig.RedirectURL)
	params.Set("state", generateState())
	params.Set("scope", "email")

	u, err := url.Parse(vkAuthConfig.AuthURL)
	if err != nil {
		return "", err
	}
	u.RawQuery = params.Encode()
	return u.String(), nil
}

func generateCodeVerifier() string {
	return "somerandomstringsomerandomstringsomerandomstringsomerandomstring"
}

func generateCodeChallenge(codeVerifier string) string {
	hash := generateHash(codeVerifier)
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func generateHash(data string) []byte {
	r := sha256.Sum256([]byte(data))
	return r[:]
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("err parsing query params: %s", err.Error())))
		return
	}
	authCode := params.Get("code")
	if authCode == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("can't get exchange code")))
		return
	}
	deviceId := params.Get("device_id")
	if deviceId == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("can't get deviceId")))
		return
	}

	tokenResp, err := ExchangeAuthCode(authCode, deviceId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("vk token exchange: %s", err.Error())))
		return
	}
	userInfo, err := FetchVKUserInfo(tokenResp.AccessToken)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("vk user info fetch: %s", err.Error())))
		return
	}
	if strconv.Itoa(tokenResp.UserId) != userInfo.User.UserId {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("vk user info id mismatch")))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("user info: %+v", userInfo.User)))
}

// ExchangeAuthCode обменивает код авторизации на токены
func ExchangeAuthCode(authCode string, deviceId string) (*AuthResp, error) {
	state := generateState()

	resp, err := http.PostForm(vkAuthConfig.TokenURL, getExchangeFormValues(authCode, deviceId, state))
	if err != nil {
		return nil, fmt.Errorf("POST request for VK token: %w", err)
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			log.Printf("close vk response body: %v", closeErr)
		}
	}()

	var vkResp AuthResp
	if err = json.NewDecoder(resp.Body).Decode(&vkResp); err != nil {
		return nil, fmt.Errorf("decoding vk resp body with tokens: %w", err)
	}
	if vkResp.Error != "" {
		return nil, fmt.Errorf("%v: %v", vkResp.Error, vkResp.ErrorDescription)
	}
	if vkResp.State != state {
		return nil, fmt.Errorf("verify state")
	}

	return &vkResp, nil
}

const stateCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"

// generateState генерирует случайную строку состояния
func generateState() string {
	state := make([]byte, 32)
	for i := range state {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(stateCharset))))
		state[i] = stateCharset[num.Int64()]
	}
	return string(state)
}

// getExchangeFormValues возвращает значения формы для запроса получения токенов
func getExchangeFormValues(authCode string, deviceId string, state string) url.Values {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code_verifier", generateCodeVerifier())
	form.Set("redirect_uri", vkAuthConfig.RedirectURL)
	form.Set("code", authCode)
	form.Set("client_id", vkAuthConfig.ClientID)
	form.Set("device_id", deviceId)
	form.Set("state", state)
	return form
}

// FetchVKUserInfo получает данные пользователя
func FetchVKUserInfo(accessToken string) (*UserInfo, error) {
	form := url.Values{}
	form.Set("access_token", accessToken)
	form.Set("client_id", vkAuthConfig.ClientID)

	resp, err := http.PostForm(vkAuthConfig.UserInfoURL, form)
	if err != nil {
		return nil, fmt.Errorf("POST request for VK user info: %w", err)
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			log.Printf("close vk response body: %v", closeErr)
		}
	}()

	var userInfo UserInfo
	if err = json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("decoding vk resp body with user info: %w", err)
	}
	if userInfo.Error != "" {
		return nil, fmt.Errorf("%v: %v", userInfo.Error, userInfo.ErrorDescription)
	}
	return &userInfo, nil
}
