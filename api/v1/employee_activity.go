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

type EmployeeActivityHandler struct {
	Repo       repository.EmployeeActivityRepository
	DeviceRepo repository.EmployeeDeviceRepository
	DB         *gorm.DB
}

type SaveEmployeeActivityRequest struct {
	KeyboardActivity int64      `json:"keyboard_activity"`
	MouseActivity    int64      `json:"mouse_activity"`
	Keystrokes       []json.RawMessage `json:"keystrokes"`
	MouseEvents      []json.RawMessage `json:"mouse_events"`
	CapturedAt       *time.Time        `json:"captured_at"`
}

func (h *EmployeeActivityHandler) AppCreate(w http.ResponseWriter, r *http.Request) {
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

	var req SaveEmployeeActivityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.KeyboardActivity < 0 || req.MouseActivity < 0 {
		jsonError(w, "keyboard_activity and mouse_activity cannot be negative", http.StatusBadRequest)
		return
	}
	keystrokes, err := json.Marshal(req.Keystrokes)
	if err != nil {
		jsonError(w, "Invalid keystrokes data", http.StatusBadRequest)
		return
	}
	mouseEvents, err := json.Marshal(req.MouseEvents)
	if err != nil {
		jsonError(w, "Invalid mouse_events data", http.StatusBadRequest)
		return
	}

	capturedAt := time.Now().UTC()
	if req.CapturedAt != nil {
		capturedAt = req.CapturedAt.UTC()
	}

	startOfDay := time.Date(capturedAt.Year(), capturedAt.Month(), capturedAt.Day(), 0, 0, 0, 0, capturedAt.Location())
	endOfDay := startOfDay.AddDate(0, 0, 1)

	var existingActivity models.EmployeeActivity
	err = h.DB.Where("employee_device_id = ? AND captured_at >= ? AND captured_at < ?", device.ID, startOfDay, endOfDay).First(&existingActivity).Error

	if err == nil {
		// Update existing record
		existingActivity.KeyboardActivity += req.KeyboardActivity
		existingActivity.MouseActivity += req.MouseActivity

		var existingKeystrokes []json.RawMessage
		json.Unmarshal(existingActivity.Keystrokes, &existingKeystrokes)
		existingKeystrokes = append(existingKeystrokes, req.Keystrokes...)
		mergedKeystrokes, _ := json.Marshal(existingKeystrokes)
		existingActivity.Keystrokes = json.RawMessage(mergedKeystrokes)

		var existingMouseEvents []json.RawMessage
		json.Unmarshal(existingActivity.MouseEvents, &existingMouseEvents)
		existingMouseEvents = append(existingMouseEvents, req.MouseEvents...)
		mergedMouseEvents, _ := json.Marshal(existingMouseEvents)
		existingActivity.MouseEvents = json.RawMessage(mergedMouseEvents)

		existingActivity.CapturedAt = capturedAt

		if updateErr := h.DB.Save(&existingActivity).Error; updateErr != nil {
			jsonError(w, "Failed to update employee activity", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"activity": existingActivity,
		})
		return
	}

	activity := &models.EmployeeActivity{
		EmployeeID:       device.EmployeeID,
		EmployeeDeviceID: &device.ID,
		KeyboardActivity: req.KeyboardActivity,
		MouseActivity:    req.MouseActivity,
		Keystrokes:       json.RawMessage(keystrokes),
		MouseEvents:      json.RawMessage(mouseEvents),
		CapturedAt:       capturedAt,
	}

	if err := h.Repo.Create(activity); err != nil {
		jsonError(w, "Failed to save employee activity", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"activity": activity,
	})
}

func (h *EmployeeActivityHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := h.DB.First(&user, middleware.GetUserID(r)).Error; err != nil {
		jsonError(w, "User record not found", http.StatusNotFound)
		return
	}
	var employee models.Employee
	if err := h.DB.Where("email = ?", user.Email).First(&employee).Error; err != nil {
		jsonError(w, "Employee record not found", http.StatusNotFound)
		return
	}
	page, limit, offset := utils.GetPaginationParams(r)
	activities, totalCount, err := h.Repo.FindByEmployeeID(employee.ID, limit, offset)
	if err != nil {
		jsonError(w, "Failed to fetch activities", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(utils.PaginatedResponse{
		Success:    true,
		Data:       activities,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: utils.CalculateTotalPages(totalCount, limit),
	})
}

func (h *EmployeeActivityHandler) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	var user models.User
	if err := h.DB.First(&user, middleware.GetUserID(r)).Error; err != nil || user.UserType == 2 {
		jsonError(w, "Admin access required", http.StatusForbidden)
		return false
	}
	return true
}

func (h *EmployeeActivityHandler) AdminList(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	employeeID, departmentID, err := activityFilters(r)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	page, limit, offset := utils.GetPaginationParams(r)
	activities, totalCount, err := h.Repo.FindAll(employeeID, departmentID, limit, offset)
	if err != nil {
		jsonError(w, "Failed to fetch activities", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(utils.PaginatedResponse{
		Success:    true,
		Data:       activities,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: utils.CalculateTotalPages(totalCount, limit),
	})
}

func (h *EmployeeActivityHandler) AdminListByEmployee(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	employeeID, err := uuid.Parse(r.PathValue("employee_id"))
	if err != nil {
		jsonError(w, "Invalid employee ID", http.StatusBadRequest)
		return
	}
	page, limit, offset := utils.GetPaginationParams(r)
	activities, totalCount, err := h.Repo.FindByEmployeeID(employeeID, limit, offset)
	if err != nil {
		jsonError(w, "Failed to fetch employee activities", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(utils.PaginatedResponse{
		Success:    true,
		Data:       activities,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: utils.CalculateTotalPages(totalCount, limit),
	})
}

func activityFilters(r *http.Request) (*uuid.UUID, *uuid.UUID, error) {
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
