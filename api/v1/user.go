package v1

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"go-api/middleware"
	"go-api/repository"
	"gorm.io/gorm"
)

type UserHandler struct {
	Repo repository.UserRepository
}

type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UpdateStatusRequest struct {
	IsBanned bool `json:"is_banned"`
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {

	userID := middleware.GetUserID(r)

	user, err := h.Repo.FindByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			jsonError(w, "User not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"user":    user,
	})
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		jsonError(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.Repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			jsonError(w, "User not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"user":    user,
	})
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {

	userID := middleware.GetUserID(r)

	var req UpdateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if req.Name == "" || req.Email == "" {
		jsonError(w, "Name and email are required", http.StatusBadRequest)
		return
	}

	user, err := h.Repo.FindByID(userID)
	if err != nil {
		jsonError(w, "User not found", http.StatusNotFound)
		return
	}

	user.Name = req.Name
	user.Email = req.Email

	if err := h.Repo.Update(user); err != nil {
		jsonError(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User updated successfully",
		"user":    user,
	})
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {

	userID := middleware.GetUserID(r)

	rowsAffected, err := h.Repo.Delete(userID)

	if err != nil {
		jsonError(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		jsonError(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User deleted successfully",
	})
}

func (h *UserHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		jsonError(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	rowsAffected, err := h.Repo.UpdateStatus(id, req.IsBanned)

	if err != nil {
		jsonError(w, "Failed to update user status", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		jsonError(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User status updated successfully",
	})
}