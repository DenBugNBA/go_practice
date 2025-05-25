package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	// подтягивает переменные окружения из .env
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/vk"
)

var (
	vkClientID     = os.Getenv("VK_CLIENT_ID")
	vkClientSecret = os.Getenv("VK_CLIENT_SECRET")
	redirectURI    = os.Getenv("VK_REDIRECT_URI") // пример: https://abc123.ngrok.io/callback

	vkOAuthConfig = &oauth2.Config{
		ClientID:     vkClientID,
		ClientSecret: vkClientSecret,
		RedirectURL:  redirectURI,
		Scopes:       []string{"email"},
		Endpoint:     vk.Endpoint,
	}
)

func main() {
	clientID := os.Getenv("VK_CLIENT_ID")
	fmt.Println("clientID:", clientID)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", indexHandler)
	r.Get("/login", loginHandler)

	fmt.Println("server is running at port 8080")
	err := http.ListenAndServe("127.0.0.1:8080", r)
	if err != nil {
		log.Err(err).Msg("server stopped with err")
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<a href="/login">Login with VK</a>`))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	url := vkOAuthConfig.AuthCodeURL("state-token") // state желательно делать динамическим
	http.Redirect(w, r, url, http.StatusFound)
}
