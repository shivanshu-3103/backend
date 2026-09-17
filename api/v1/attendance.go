package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go-api/models"
	"go-api/repository"
	"go-api/utils"
	"go-api/middleware"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AttendanceHandler struct {
	Repo repository.AttendanceRepository
	DB   *gorm.DB
}

func (h *AttendanceHandler) employeeFromRequest(r *http.Request) (*models.Employee, error) {
	var user models.User
	if err := h.DB.First(&user, middleware.GetUserID(r)).Error; err != nil {
		return nil, fmt.Errorf("Unauthorized")
	}
	var employee models.Employee
	err := h.DB.Where("email = ?", user.Email).First(&employee).Error
	if err != nil {
		return nil, fmt.Errorf("Unauthorized")
	}
	return &employee, nil
}

func (h *AttendanceHandler) getOrCreateTodayAttendance(employeeID string) (*models.Attendance, error) {
    // Utility to parse employee ID
    parsedID, err := uuid.Parse(employeeID)
    if err != nil {
        return nil, err
    }

	now := time.Now()
	// Just use today's date (midnight)
	todayStr := now.Format("2006-01-02")
	today, _ := time.Parse("2006-01-02", todayStr)

	att, err := h.Repo.FindTodayByEmployeeID(parsedID, today)
	if err != nil {
		// Create new if not found
		att = &models.Attendance{
			EmployeeID: parsedID,
			Date:       today,
			Status:     "not_started",
		}
		if createErr := h.Repo.Create(att); createErr != nil {
			return nil, createErr
		}
	}
	return att, nil
}

func (h *AttendanceHandler) ClockIn(w http.ResponseWriter, r *http.Request) {
	employee, err := h.employeeFromRequest(r)
	if err != nil {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		WorkType string `json:"work_type"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.WorkType == "" {
		req.WorkType = "office"
	}

	att, err := h.getOrCreateTodayAttendance(employee.ID.String())
	if err != nil {
		jsonError(w, "Failed to get attendance record", http.StatusInternalServerError)
		return
	}

	now := time.Now()

	if att.FirstClockIn == nil {
		att.FirstClockIn = &now
	}
	att.Status = "working"
	att.WorkType = req.WorkType
	h.Repo.Update(att)

	event := &models.AttendanceEvent{
		AttendanceID: att.ID,
		EventType:    "clock_in",
		EventTime:    now,
	}
	h.Repo.CreateEvent(event)

	utils.LogSuccess("Employee %s clocked in", employee.ID)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "attendance": att})
}

func (h *AttendanceHandler) ClockOut(w http.ResponseWriter, r *http.Request) {
	employee, err := h.employeeFromRequest(r)
	if err != nil {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	att, err := h.getOrCreateTodayAttendance(employee.ID.String())
	if err != nil || att.FirstClockIn == nil {
		jsonError(w, "No active attendance record found", http.StatusBadRequest)
		return
	}

	now := time.Now()
	att.LastClockOut = &now
	att.Status = "completed"

	// Calculate total work secs roughly (should account for breaks, but this is a simple example)
	if att.FirstClockIn != nil {
	    // simplistic total time calculation
	    duration := now.Sub(*att.FirstClockIn)
	    att.TotalWorkSecs = int(duration.Seconds()) - att.TotalBreakSecs
	}

	h.Repo.Update(att)

	event := &models.AttendanceEvent{
		AttendanceID: att.ID,
		EventType:    "clock_out",
		EventTime:    now,
	}
	h.Repo.CreateEvent(event)

	utils.LogSuccess("Employee %s clocked out", employee.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "attendance": att})
}

func (h *AttendanceHandler) BreakStart(w http.ResponseWriter, r *http.Request) {
	employee, err := h.employeeFromRequest(r)
	if err != nil {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Reason      string `json:"reason"`
		OtherReason string `json:"other_reason"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	att, err := h.getOrCreateTodayAttendance(employee.ID.String())
	if err != nil || att.FirstClockIn == nil {
		jsonError(w, "Cannot take a break without clocking in first", http.StatusBadRequest)
		return
	}

	att.Status = "on_break"
	h.Repo.Update(att)

	now := time.Now()
	event := &models.AttendanceEvent{
		AttendanceID: att.ID,
		EventType:    "break_start",
		EventTime:    now,
		Reason:       req.Reason,
		OtherReason:  req.OtherReason,
	}
	h.Repo.CreateEvent(event)

	utils.LogSuccess("Employee %s took a break", employee.ID)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "attendance": att})
}

func (h *AttendanceHandler) BreakEnd(w http.ResponseWriter, r *http.Request) {
	employee, err := h.employeeFromRequest(r)
	if err != nil {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	att, err := h.getOrCreateTodayAttendance(employee.ID.String())
	if err != nil || att.Status != "on_break" {
		jsonError(w, "Not currently on break", http.StatusBadRequest)
		return
	}

	now := time.Now()
	
	// Find the last break_start event to calculate break duration
	// This is simplified. In a real app we'd fetch the last event from the DB.
	// For now we'll just log the end event and change status.
	
	att.Status = "working"
	h.Repo.Update(att)

	event := &models.AttendanceEvent{
		AttendanceID: att.ID,
		EventType:    "break_end",
		EventTime:    now,
	}
	h.Repo.CreateEvent(event)

	utils.LogSuccess("Employee %s ended their break", employee.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "attendance": att})
}

// GetTimesheet for the logged-in employee
func (h *AttendanceHandler) GetTimesheet(w http.ResponseWriter, r *http.Request) {
	employee, err := h.employeeFromRequest(r)
	if err != nil {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	page, limit, offset := utils.GetPaginationParams(r)
	attendances, total, err := h.Repo.FindAllByEmployeeID(employee.ID, limit, offset)
	if err != nil {
		jsonError(w, "Failed to fetch timesheet", http.StatusInternalServerError)
		return
	}

	response := utils.PaginatedResponse{
		Success:    true,
		Data:       attendances,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
		TotalPages: utils.CalculateTotalPages(total, limit),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Admin API
func (h *AttendanceHandler) AdminGetAllTimesheets(w http.ResponseWriter, r *http.Request) {
	page, limit, offset := utils.GetPaginationParams(r)
	attendances, total, err := h.Repo.FindAll(limit, offset)
	if err != nil {
		jsonError(w, "Failed to fetch timesheets", http.StatusInternalServerError)
		return
	}

	response := utils.PaginatedResponse{
		Success:    true,
		Data:       attendances,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
		TotalPages: utils.CalculateTotalPages(total, limit),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *AttendanceHandler) AdminGetEmployeeTimesheet(w http.ResponseWriter, r *http.Request) {
	employeeIDStr := r.PathValue("id")
	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		jsonError(w, "Invalid employee ID", http.StatusBadRequest)
		return
	}

	page, limit, offset := utils.GetPaginationParams(r)
	attendances, total, err := h.Repo.FindAllByEmployeeID(employeeID, limit, offset)
	if err != nil {
		jsonError(w, "Failed to fetch timesheets", http.StatusInternalServerError)
		return
	}

	response := utils.PaginatedResponse{
		Success:    true,
		Data:       attendances,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
		TotalPages: utils.CalculateTotalPages(total, limit),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
