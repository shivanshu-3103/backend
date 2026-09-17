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

type EmployeeDeviceHandler struct {
	DeviceRepo   repository.EmployeeDeviceRepository
	EmployeeRepo repository.EmployeeRepository
	DB           *gorm.DB // Needed to lookup employee_id from user email
}

type RegisterDeviceRequest struct {
	DeviceFingerprint     string `json:"device_fingerprint"`
	MachineGuidHash       string `json:"machine_guid_hash"`
	BiosSerialHash        string `json:"bios_serial_hash"`
	MotherboardSerialHash string `json:"motherboard_serial_hash"`
	DeviceName            string `json:"device_name"`
	Manufacturer          string `json:"manufacturer"`
	Model                 string `json:"model"`
	OSName                string `json:"os_name"`
	OSVersion             string `json:"os_version"`
	Architecture          string `json:"architecture"`
	CPUName               string `json:"cpu_name"`
	RamTotalMB            int64  `json:"ram_total_mb"`
	DiskTotalGB           int64  `json:"disk_total_gb"`
	AppVersion            string `json:"app_version"`
}

// POST /api/v1/employee-devices/register
func (h *EmployeeDeviceHandler) Register(w http.ResponseWriter, r *http.Request) {
	// The user is authenticated via JWT.
	// We get the user email from the token context to find their employee record
	userID := middleware.GetUserID(r)

	var user models.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		jsonError(w, "User record not found", http.StatusNotFound)
		return
	}

	var employee models.Employee
	if err := h.DB.Where("email = ?", user.Email).First(&employee).Error; err != nil {
		jsonError(w, "Employee record not found for this user", http.StatusNotFound)
		return
	}

	var req RegisterDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.DeviceFingerprint == "" {
		jsonError(w, "device_fingerprint is required", http.StatusBadRequest)
		return
	}

	now := time.Now()

	device := models.EmployeeDevice{
		EmployeeID:            employee.ID,
		DeviceFingerprint:     req.DeviceFingerprint,
		MachineGuidHash:       req.MachineGuidHash,
		BiosSerialHash:        req.BiosSerialHash,
		MotherboardSerialHash: req.MotherboardSerialHash,
		DeviceName:            req.DeviceName,
		Manufacturer:          req.Manufacturer,
		Model:                 req.Model,
		OSName:                req.OSName,
		OSVersion:             req.OSVersion,
		Architecture:          req.Architecture,
		CPUName:               req.CPUName,
		RamTotalMB:            req.RamTotalMB,
		DiskTotalGB:           req.DiskTotalGB,
		AppVersion:            req.AppVersion,
		Status:                "active",
		LastSeenAt:            &now,
	}

	if err := h.DeviceRepo.UpsertDevice(&device); err != nil {
		jsonError(w, "Failed to register or update device", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Device registered successfully",
		"device":  device,
	})
}

// GET /api/v1/employee-devices
func (h *EmployeeDeviceHandler) ListMyDevices(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var user models.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		jsonError(w, "User record not found", http.StatusNotFound)
		return
	}

	var employee models.Employee
	if err := h.DB.Where("email = ?", user.Email).First(&employee).Error; err != nil {
		jsonError(w, "Employee record not found", http.StatusNotFound)
		return
	}

	page, limit, offset := utils.GetPaginationParams(r)
	devices, totalCount, err := h.DeviceRepo.FindByEmployeeID(employee.ID, limit, offset)
	if err != nil {
		jsonError(w, "Failed to fetch devices", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(utils.PaginatedResponse{
		Success:    true,
		Data:       devices,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: utils.CalculateTotalPages(totalCount, limit),
	})
}

// DELETE /api/v1/employee-devices/{id}
func (h *EmployeeDeviceHandler) DeleteDevice(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var user models.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		jsonError(w, "User record not found", http.StatusNotFound)
		return
	}

	var employee models.Employee
	if err := h.DB.Where("email = ?", user.Email).First(&employee).Error; err != nil {
		jsonError(w, "Employee record not found", http.StatusNotFound)
		return
	}

	idStr := r.PathValue("id")
	deviceID, err := uuid.Parse(idStr)
	if err != nil {
		jsonError(w, "Invalid device ID", http.StatusBadRequest)
		return
	}

	// Verify the device belongs to this employee
	device, err := h.DeviceRepo.FindByID(deviceID)
	if err != nil {
		jsonError(w, "Device not found", http.StatusNotFound)
		return
	}

	if device.EmployeeID != employee.ID {
		jsonError(w, "Unauthorized to delete this device", http.StatusForbidden)
		return
	}

	if err := h.DeviceRepo.DeleteDevice(deviceID); err != nil {
		jsonError(w, "Failed to delete device", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Device deleted successfully",
	})
}

// GET /api/v1/admin/employee-devices
func (h *EmployeeDeviceHandler) ListAllDevices(w http.ResponseWriter, r *http.Request) {
	page, limit, offset := utils.GetPaginationParams(r)
	devices, totalCount, err := h.DeviceRepo.FindAll(limit, offset)
	if err != nil {
		jsonError(w, "Failed to fetch devices", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(utils.PaginatedResponse{
		Success:    true,
		Data:       devices,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: utils.CalculateTotalPages(totalCount, limit),
	})
}

// GET /api/v1/admin/employees/{id}/devices
func (h *EmployeeDeviceHandler) ListDevicesByEmployee(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	employeeID, err := uuid.Parse(idStr)
	if err != nil {
		jsonError(w, "Invalid employee ID", http.StatusBadRequest)
		return
	}

	page, limit, offset := utils.GetPaginationParams(r)
	devices, totalCount, err := h.DeviceRepo.FindByEmployeeID(employeeID, limit, offset)
	if err != nil {
		jsonError(w, "Failed to fetch devices", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(utils.PaginatedResponse{
		Success:    true,
		Data:       devices,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: utils.CalculateTotalPages(totalCount, limit),
	})
}

// GET /api/v1/app/device/check
func (h *EmployeeDeviceHandler) CheckDevice(w http.ResponseWriter, r *http.Request) {
	deviceFingerprint := r.URL.Query().Get("device_fingerprint")
	if deviceFingerprint == "" {
		jsonError(w, "device_fingerprint is required", http.StatusBadRequest)
		return
	}

	device, err := h.DeviceRepo.FindByFingerprint(deviceFingerprint)
	if err != nil {
		jsonError(w, "Device not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Device is registered",
		"device_id": device.ID,
		"employee_id": device.EmployeeID,
	})
}
