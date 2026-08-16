package handler

import (
	"IssueForge/internal/dto"
	"IssueForge/internal/httpx"
	"IssueForge/internal/middleware"
	"IssueForge/internal/repository"
	"IssueForge/internal/service"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

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
