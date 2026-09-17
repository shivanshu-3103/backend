package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	"go-api/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OptionsHandler struct {
	DB *gorm.DB
}

type OptionItem struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type TaskOptionItem struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Type string    `json:"type"`
}

func (h *OptionsHandler) getOptions(w http.ResponseWriter, r *http.Request, model interface{}, where ...interface{}) {
	var options []OptionItem

	query := h.DB.Model(model).Select("id, name")
	if len(where) > 0 {
		query = query.Where(where[0], where[1:]...)
	}

	if err := query.Find(&options).Error; err != nil {
		jsonError(w, "Failed to fetch options", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    options,
	})
}

func (h *OptionsHandler) Departments(w http.ResponseWriter, r *http.Request) {
	h.getOptions(w, r, &models.Department{}, "status = ?", "active")
}

func (h *OptionsHandler) Roles(w http.ResponseWriter, r *http.Request) {
	h.getOptions(w, r, &models.Role{}, "status = ?", "active")
}

func (h *OptionsHandler) Teams(w http.ResponseWriter, r *http.Request) {
	h.getOptions(w, r, &models.Team{})
}

func (h *OptionsHandler) Tasks(w http.ResponseWriter, r *http.Request) {
	taskType := r.URL.Query().Get("type")

	var options []TaskOptionItem
	query := h.DB.Model(&models.Task{}).Select("id, name, type").Where("status = ?", "active")

	if taskType != "" {
		types := strings.Split(taskType, ",")
		query = query.Where("type IN ?", types)
	}

	if err := query.Find(&options).Error; err != nil {
		jsonError(w, "Failed to fetch task options", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    options,
	})
}

func (h *OptionsHandler) Employees(w http.ResponseWriter, r *http.Request) {
	h.getOptions(w, r, &models.Employee{}, "status = ?", "active")
}
