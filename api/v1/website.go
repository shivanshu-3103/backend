package v1

import (
	"encoding/json"
	"net/http"
	"time"

	"go-api/middleware"
	"go-api/models"
	"go-api/repository"
	"go-api/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WebsiteHandler struct {
	Repo       repository.WebsiteRepository
	DeviceRepo repository.EmployeeDeviceRepository
	DB         *gorm.DB
}

type SaveWebsiteRequest struct {
	Domain          string     `json:"domain"`
	URL             string     `json:"url"`
	PageTitle       string     `json:"page_title"`
	StartedAt       time.Time  `json:"started_at"`
	EndedAt         *time.Time `json:"ended_at"`
	DurationSeconds int64      `json:"duration_seconds"`
}

func (h *WebsiteHandler) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	var user models.User
	if err := h.DB.First(&user, middleware.GetUserID(r)).Error; err != nil || user.UserType == 2 {
		jsonError(w, "Admin access required", http.StatusForbidden)
		return false
	}
	return true
}

func (h *WebsiteHandler) AppCreate(w http.ResponseWriter, r *http.Request) {
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

	var req SaveWebsiteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Domain == "" || req.URL == "" {
		jsonError(w, "domain and url are required", http.StatusBadRequest)
		return
	}

	var endedAt time.Time
	if req.EndedAt != nil {
		endedAt = *req.EndedAt
	}

	website := &models.Website{
		EmployeeID:      device.EmployeeID,
		Domain:          req.Domain,
		URL:             req.URL,
		PageTitle:       req.PageTitle,
		StartedAt:       req.StartedAt,
		EndedAt:         endedAt,
		DurationSeconds: req.DurationSeconds,
	}

	if err := h.Repo.Create(website); err != nil {
		jsonError(w, "Failed to save website data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"website": website,
	})
}

func (h *WebsiteHandler) AdminList(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	employeeID, departmentID, err := parseFilters(r)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	page, limit, offset := utils.GetPaginationParams(r)
	websites, totalCount, err := h.Repo.FindAll(employeeID, departmentID, limit, offset)
	if err != nil {
		jsonError(w, "Failed to fetch websites", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(utils.PaginatedResponse{
		Success:    true,
		Data:       websites,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: utils.CalculateTotalPages(totalCount, limit),
	})
}

func (h *WebsiteHandler) AdminListByEmployee(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	employeeID, err := uuid.Parse(r.PathValue("employee_id"))
	if err != nil {
		jsonError(w, "Invalid employee ID", http.StatusBadRequest)
		return
	}
	page, limit, offset := utils.GetPaginationParams(r)
	websites, totalCount, err := h.Repo.FindByEmployeeID(employeeID, limit, offset)
	if err != nil {
		jsonError(w, "Failed to fetch employee websites", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(utils.PaginatedResponse{
		Success:    true,
		Data:       websites,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: utils.CalculateTotalPages(totalCount, limit),
	})
}

func (h *WebsiteHandler) AdminUpdate(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		jsonError(w, "Invalid website ID", http.StatusBadRequest)
		return
	}
	
	website, err := h.Repo.FindByID(id)
	if err != nil {
		jsonError(w, "Website not found", http.StatusNotFound)
		return
	}

	var req SaveWebsiteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Domain != "" {
		website.Domain = req.Domain
	}
	if req.URL != "" {
		website.URL = req.URL
	}
	if req.PageTitle != "" {
		website.PageTitle = req.PageTitle
	}
	if !req.StartedAt.IsZero() {
		website.StartedAt = req.StartedAt
	}
	if req.EndedAt != nil {
		website.EndedAt = *req.EndedAt
	}
	if req.DurationSeconds > 0 {
		website.DurationSeconds = req.DurationSeconds
	}

	if err := h.Repo.Update(website); err != nil {
		jsonError(w, "Failed to update website", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"website": website,
	})
}

func (h *WebsiteHandler) AdminDelete(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		jsonError(w, "Invalid website ID", http.StatusBadRequest)
		return
	}
	if err := h.Repo.Delete(id); err != nil {
		jsonError(w, "Failed to delete website", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Website deleted successfully",
	})
}
