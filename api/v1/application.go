package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go-api/middleware"
	"go-api/models"
	"go-api/repository"
	"go-api/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ApplicationHandler struct {
	Repo       repository.ApplicationRepository
	DeviceRepo repository.EmployeeDeviceRepository
	DB         *gorm.DB
}

type SaveApplicationRequest struct {
	ApplicationName string     `json:"application_name"`
	ExecutableName  string     `json:"executable_name"`
	WindowTitle     string     `json:"window_title"`
	StartedAt       time.Time  `json:"started_at"`
	EndedAt         *time.Time `json:"ended_at"`
	DurationSeconds int64      `json:"duration_seconds"`
}

func (h *ApplicationHandler) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	var user models.User
	if err := h.DB.First(&user, middleware.GetUserID(r)).Error; err != nil || user.UserType == 2 {
		jsonError(w, "Admin access required", http.StatusForbidden)
		return false
	}
	return true
}

func (h *ApplicationHandler) AppCreate(w http.ResponseWriter, r *http.Request) {
	deviceFingerprint := r.Header.Get("X-Device-Fingerprint")
	if deviceFingerprint == "" {
		jsonError(w, "X-Device-Fingerprint header is required", http.StatusUnauthorized)
		return
	}

	device, err := h.DeviceRepo.FindByFingerprint(deviceFingerprint)
	if err != nil {
		jsonError(w, "Device not registered", http.StatusUnauthorized)
		return
	}

	var req SaveApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ApplicationName == "" || req.ExecutableName == "" {
		jsonError(w, "application_name and executable_name are required", http.StatusBadRequest)
		return
	}

	var endedAt time.Time
	if req.EndedAt != nil {
		endedAt = *req.EndedAt
	}

	app := &models.Application{
		EmployeeID:      device.EmployeeID,
		ApplicationName: req.ApplicationName,
		ExecutableName:  req.ExecutableName,
		WindowTitle:     req.WindowTitle,
		StartedAt:       req.StartedAt,
		EndedAt:         endedAt,
		DurationSeconds: req.DurationSeconds,
	}

	if err := h.Repo.Create(app); err != nil {
		jsonError(w, "Failed to save application data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"application": app,
	})
}

func (h *ApplicationHandler) AdminList(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	employeeID, departmentID, err := parseFilters(r)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	page, limit, offset := utils.GetPaginationParams(r)
	apps, totalCount, err := h.Repo.FindAll(employeeID, departmentID, limit, offset)
	if err != nil {
		jsonError(w, "Failed to fetch applications", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(utils.PaginatedResponse{
		Success:    true,
		Data:       apps,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: utils.CalculateTotalPages(totalCount, limit),
	})
}

func (h *ApplicationHandler) AdminListByEmployee(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	employeeID, err := uuid.Parse(r.PathValue("employee_id"))
	if err != nil {
		jsonError(w, "Invalid employee ID", http.StatusBadRequest)
		return
	}
	page, limit, offset := utils.GetPaginationParams(r)
	apps, totalCount, err := h.Repo.FindByEmployeeID(employeeID, limit, offset)
	if err != nil {
		jsonError(w, "Failed to fetch employee applications", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(utils.PaginatedResponse{
		Success:    true,
		Data:       apps,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: utils.CalculateTotalPages(totalCount, limit),
	})
}

func (h *ApplicationHandler) AdminUpdate(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		jsonError(w, "Invalid application ID", http.StatusBadRequest)
		return
	}
	
	app, err := h.Repo.FindByID(id)
	if err != nil {
		jsonError(w, "Application not found", http.StatusNotFound)
		return
	}

	var req SaveApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ApplicationName != "" {
		app.ApplicationName = req.ApplicationName
	}
	if req.ExecutableName != "" {
		app.ExecutableName = req.ExecutableName
	}
	if req.WindowTitle != "" {
		app.WindowTitle = req.WindowTitle
	}
	if !req.StartedAt.IsZero() {
		app.StartedAt = req.StartedAt
	}
	if req.EndedAt != nil {
		app.EndedAt = *req.EndedAt
	}
	if req.DurationSeconds > 0 {
		app.DurationSeconds = req.DurationSeconds
	}

	if err := h.Repo.Update(app); err != nil {
		jsonError(w, "Failed to update application", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"application": app,
	})
}

func (h *ApplicationHandler) AdminDelete(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		jsonError(w, "Invalid application ID", http.StatusBadRequest)
		return
	}
	if err := h.Repo.Delete(id); err != nil {
		jsonError(w, "Failed to delete application", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Application deleted successfully",
	})
}

func parseFilters(r *http.Request) (*uuid.UUID, *uuid.UUID, error) {
	parse := func(name string) (*uuid.UUID, error) {
		value := r.URL.Query().Get(name)
		if value == "" {
			return nil, nil
		}
		id, err := uuid.Parse(value)
		return &id, err
	}
	employeeID, err := parse("employee_id")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid employee_id")
	}
	departmentID, err := parse("department_id")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid department_id")
	}
	return employeeID, departmentID, nil
}
