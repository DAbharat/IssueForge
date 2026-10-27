package handler

import (
	"IssueForge/internal/dto"
	"IssueForge/internal/httpx"
	"IssueForge/internal/middleware"
	"IssueForge/internal/repository"
	"IssueForge/internal/service"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type UserService interface {
	CreateUser(ctx context.Context, req dto.CreateUserRequest) (dto.CreateUserResponse, error)
	Login(ctx context.Context, req dto.LoginUserRequest) (dto.LoginUserResponse, error)
	GetCurrentUser(ctx context.Context, userID int64) (dto.MeResponse, error)
	RefreshAccessToken(ctx context.Context, req dto.RefreshTokenRequest) (string, error)
	GetUserByID(ctx context.Context, id int64) (dto.UserResponse, error)
	GetUserByUsername(ctx context.Context, username string) (dto.UserResponse, error)
	SearchUserByUsername(ctx context.Context, search *string) ([]dto.UserResponse, error)
	DeleteUser(ctx context.Context, id int64) (int64, error)
	GoogleAuthURL(state string) (string, error)
	GoogleLogin(ctx context.Context, code string) (dto.LoginUserResponse, error)
	Logout(ctx context.Context, token string) error
	CreateHandoff(ctx context.Context, accessToken, refreshToken string) (string, error)
	ExchangeHandoff(ctx context.Context, code string) (dto.ExchangeResponse, error)
}

type UserHandler struct {
	userService UserService
}

func NewUserHandler(service UserService) *UserHandler {
	return &UserHandler{
		userService: service,
	}
}

func (h *UserHandler) Signup(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	var req dto.CreateUserRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		log.Printf("decode error: %v", err)
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid or oversized request body")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		httpx.RespondWithError(w, http.StatusBadRequest, "request body must contain a single JSON object")
		return
	}

	user, err := h.userService.CreateUser(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPassword),
			errors.Is(err, service.ErrInvalidEmail),
			errors.Is(err, service.ErrInvalidFullName):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, repository.ErrDuplicateEmail):
			httpx.RespondWithError(w, http.StatusConflict, err.Error())
		default:
			log.Printf("user signup fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusCreated, user)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	var req dto.LoginUserRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		httpx.RespondWithError(w, http.StatusBadRequest, "request body must contain a single JSON object")
		return
	}

	user, err := h.userService.Login(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			httpx.RespondWithError(w, http.StatusUnauthorized, err.Error())
		default:
			log.Printf("user login fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, user)
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.userService.GetCurrentUser(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("me response fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, user)
}

func (h *UserHandler) RefreshAccessToken(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshTokenRequest

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid or oversized request body")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		httpx.RespondWithError(w, http.StatusBadRequest, "request body must only contain a single json object")
		return
	}

	if req.RefreshToken == "" {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid token")
		return
	}

	accessToken, err := h.userService.RefreshAccessToken(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRefTokenNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("refresh access token fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, map[string]string{
		"accessToken": accessToken,
	})
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)

	userID, err := strconv.ParseInt(vars["userID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidUserID.Error())
		return
	}

	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidUserID):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrUserNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("get user by id fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, user)
}

func (h *UserHandler) GetUserByUsername(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	username := mux.Vars(r)["username"]
	if username == "" {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidUsername.Error())
		return
	}

	user, err := h.userService.GetUserByUsername(r.Context(), username)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidUsername):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrUserNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("get user by username: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, user)
}

func (h *UserHandler) SearchUserByUsername(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	search := r.URL.Query().Get("search")

	var searchPtr *string
	if search != "" {
		searchPtr = &search
	}

	user, err := h.userService.SearchUserByUsername(r.Context(), searchPtr)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidSearchQuery):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		default:
			log.Printf("search user by username fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, user)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)

	userID, err := strconv.ParseInt(vars["userID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidUserID.Error())
		return
	}

	user, err := h.userService.DeleteUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		}
		log.Printf("delete user fail: %v", err)
		httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, user)
}

func (h *UserHandler) GoogleLoginStart(w http.ResponseWriter, r *http.Request) {
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	state := base64.RawURLEncoding.EncodeToString(stateBytes)

	isSecure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"

	http.SetCookie(w, &http.Cookie{
		Name:     "google_oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   300,
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})

	url, err := h.userService.GoogleAuthURL(state)
	if err != nil {
		log.Printf("start google login fail: %v", err)
		httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

func (h *UserHandler) GoogleLoginCallback(w http.ResponseWriter, r *http.Request) {
	isSecure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"

	http.SetCookie(w, &http.Cookie{
		Name:     "google_oauth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})

	stateCookie, err := r.Cookie("google_oauth_state")
	state := r.URL.Query().Get("state")

	if err != nil || stateCookie.Value == "" {
		log.Printf("GoogleLoginCallback: state mismatch or missing cookie: err=%v, stateParam=%s", err, state)
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid oauth state")
		return
	}
	if state == "" || state != stateCookie.Value {
		log.Printf("GoogleLoginCallback: state mismatch or missing cookie: err=%v, stateParam=%s", err, state)
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid oauth state")
	}

	if providerErr := r.URL.Query().Get("error"); providerErr != "" {
		httpx.RespondWithError(w, http.StatusUnauthorized, "google sign-in was not completed")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		httpx.RespondWithError(w, http.StatusBadRequest, "missing authorization code")
		return
	}

	result, err := h.userService.GoogleLogin(r.Context(), code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrDuplicateEmail),
			errors.Is(err, service.ErrDuplicateUsername):
			httpx.RespondWithError(w, http.StatusConflict, err.Error())
		default:
			log.Printf("google login callback fail: %v", err)
			httpx.RespondWithError(w, http.StatusUnauthorized, "google sign-in failed")
		}
		return
	}

	if r.Header.Get("Sec-Fetch-Mode") == "navigate" {
		handoffCode, err := h.userService.CreateHandoff(r.Context(), result.AccessToken, result.RefreshToken)
		if err != nil {
			log.Printf("create handoff: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:4000"
		}
		redirectTarget := fmt.Sprintf("%s/auth/callback?code=%s",
			strings.TrimRight(frontendURL, "/"),
			url.QueryEscape(handoffCode),
		)
		http.Redirect(w, r, redirectTarget, http.StatusFound)
		return
	}

	httpx.RespondWithJSON(w, http.StatusOK, result)
}

func (h *UserHandler) ExchangeHandoff(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)

	var req dto.ExchangeRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid or oversized request body")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		httpx.RespondWithError(w, http.StatusBadRequest, "request body must only contain a single json object")
		return
	}

	response, err := h.userService.ExchangeHandoff(r.Context(), req.Code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidHandoffCode):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		default:
			log.Printf("exchange handoff: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	httpx.RespondWithJSON(w, http.StatusOK, response)
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	var req dto.LogoutRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid or oversized request body")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		httpx.RespondWithError(w, http.StatusBadRequest, "request body must contain a single json object")
		return
	}

	if err := h.userService.Logout(r.Context(), req.RefreshToken); err != nil {
		log.Printf("logout fail: %v", err)
	}

	httpx.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "user logged out",
	})
}
