package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"topikkhusus-methodtracker/internal/models"
	"topikkhusus-methodtracker/internal/services"
)

var emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

type successResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

type errorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type UserHandler struct {
	service services.UserService
}

func NewUserHandler(service services.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var request models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	request.Name = strings.TrimSpace(request.Name)
	request.Email = strings.TrimSpace(request.Email)

	if request.Name == "" || request.Email == "" || !emailRegex.MatchString(request.Email) {
		writeError(w, http.StatusBadRequest, "name and valid email are required")
		return
	}

	user, err := h.service.CreateUser(r.Context(), request)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeSuccess(w, http.StatusCreated, user)
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	users, err := h.service.GetAllUsers(r.Context())
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, users)
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user, err := h.service.GetUserByID(r.Context(), id)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, user)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if err := h.service.DeleteUser(r.Context(), id); err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"message": "user deleted"})
}

func (h *UserHandler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, services.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeSuccess(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := successResponse{Success: true, Data: data}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode success response: %v", err), http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := errorResponse{Success: false, Error: message}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode error response: %v", err), http.StatusInternalServerError)
	}
}
