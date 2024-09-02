package handler

import (
	"fmt"
	"net/http"

	"github.com/DarRo9/Test-task-BackDev/internal/config"
)

const (
	name  = "Name"
	token = "Token"
)

type Auth interface {
	RefreshToken(userName string) (string, error)
	MakeAccessToken(userName string) (string, error)
	TakeValidToken(refreshToken string, userName string) bool
	CheckCountTokens(userName string) error
	InsertToken(refreshToken string, userName string) error
	SelectToken(CreateObjectToken string, userName string) error
}
type response struct {
	Name         string `json:"user_name"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Handler struct {
	config *config.Config
	auth   Auth
	logger Logger
}

type Logger func(http.Handler) http.Handler

func CreateObject(config *config.Config, auth Auth, logger Logger) *Handler {
	return &Handler{
		config: config,
		auth:   auth,
		logger: logger,
	}
}

func (h *Handler) CreateObjectRouter() http.Handler {
	router := http.NewServeMux()

	authHandler := h.logger(h.authHandler())
	router.Handle("/auth", authHandler)

	refreshHandler := h.logger(h.refreshHandler())
	router.Handle("/refresh", refreshHandler)

	return router
}

func (h *Handler) refreshHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "This method is not allowed", http.StatusMethodNotAllowed)
			return
		}

		refreshTokenFromHeader, err := getHeader(r, token)
		if err != nil {
			http.Error(w, fmt.Sprintf("Header '%v' is missing", token), http.StatusBadRequest)
			return
		}

		userName, err := getHeader(r, name)
		if err != nil {
			http.Error(w, fmt.Sprintf("Header '%v' is missing", name), http.StatusBadRequest)
			return
		}

		if ok := h.auth.TakeValidToken(refreshTokenFromHeader, userName); !ok {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		CreateObjectRefreshToken, err := h.auth.RefreshToken(userName)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err := h.auth.SelectToken(CreateObjectRefreshToken, userName); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		accessToken, err := h.auth.MakeAccessToken(userName)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		setCookies(w, CreateObjectRefreshToken, accessToken, h.config.RefreshTokenTTL, h.config.AccessTokenTTL)

		response := response{
			Name:         userName,
			AccessToken:  accessToken,
			RefreshToken: CreateObjectRefreshToken,
		}

		if err := jsonRendering(w, response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}

func (h *Handler) authHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "This method is not allowed", http.StatusMethodNotAllowed)
			return
		}

		userName, err := getHeader(r, name)
		if err != nil {
			http.Error(w, fmt.Sprintf("Header '%v' is missing", name), http.StatusBadRequest)
			return
		}

		if err := h.auth.CheckCountTokens(userName); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		refreshToken, err := h.auth.RefreshToken(userName)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err := h.auth.InsertToken(refreshToken, userName); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		accessToken, err := h.auth.MakeAccessToken(userName)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		setCookies(w, refreshToken, accessToken, h.config.RefreshTokenTTL, h.config.AccessTokenTTL)

		response := response{
			Name:         userName,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}

		if err := jsonRendering(w, response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}
