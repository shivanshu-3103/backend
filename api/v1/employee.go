package v1

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"go-api/models"
	"go-api/repository"
	"go-api/services"
	"go-api/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type EmployeeHandler struct {
	Repo         repository.EmployeeRepository
	EmailService services.EmailService
}

type CreateEmployeeRequest struct {
	OrganizationID         string `json:"organization_id"`
	EmployeeCode           string `json:"employee_code"`
	Name                   string `json:"name"`
	Email                  string `json:"email"`
	Password               string `json:"password"`
	Phone                  string `json:"phone"`
	DateOfJoining          string `json:"date_of_joining"`
	Designation            string `json:"designation"`
	DepartmentID           string `json:"department_id"`
	RoleID                 string `json:"role_id"`
	TeamID                 string `json:"team_id"`
	ReportingManager       string `json:"reporting_manager"`
	WorkLocation           string `json:"work_location"`
	EmploymentType         string `json:"employment_type"`
	Photo                  string `json:"photo"`
	Address                string `json:"address"`
	DateOfBirth            string `json:"date_of_birth"`
	Gender                 string `json:"gender"`
	BloodGroup             string `json:"blood_group"`
	EmergencyContactName   string `json:"emergency_contact_name"`
	EmergencyContactNumber string `json:"emergency_contact_number"`
	Relationship           string `json:"relationship"`
	Notes                  string `json:"notes"`
	Status                 string `json:"status"`
}

type UpdateEmployeeRequest struct {
	OrganizationID         string `json:"organization_id"`
	EmployeeCode           string `json:"employee_code"`
	Name                   string `json:"name"`
	Email                  string `json:"email"`
	Phone                  string `json:"phone"`
	DateOfJoining          string `json:"date_of_joining"`
	Designation            string `json:"designation"`
	DepartmentID           string `json:"department_id"`
	RoleID                 string `json:"role_id"`
	TeamID                 string `json:"team_id"`
	ReportingManager       string `json:"reporting_manager"`
	WorkLocation           string `json:"work_location"`
	EmploymentType         string `json:"employment_type"`
	Photo                  string `json:"photo"`
	Address                string `json:"address"`
	DateOfBirth            string `json:"date_of_birth"`
	Gender                 string `json:"gender"`
	BloodGroup             string `json:"blood_group"`
	EmergencyContactName   string `json:"emergency_contact_name"`
	EmergencyContactNumber string `json:"emergency_contact_number"`
	Relationship           string `json:"relationship"`
	Notes                  string `json:"notes"`
	Status                 string `json:"status"`
}

// POST /api/v1/employees
func (h *EmployeeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if req.Name == "" || req.Email == "" || req.EmployeeCode == "" {
		jsonError(w, "Name, email, and employee_code are required", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		max := big.NewInt(100000000)
		n, _ := rand.Int(rand.Reader, max)
		req.Password = fmt.Sprintf("%08d", n.Int64())
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		jsonError(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	status := req.Status
	if status == "" {
		status = "active"
	}

	orgID, _ := uuid.Parse(req.OrganizationID)
	deptID, _ := uuid.Parse(req.DepartmentID)
	roleID, _ := uuid.Parse(req.RoleID)
	
	var teamID *uuid.UUID
	if req.TeamID != "" {
		if id, err := uuid.Parse(req.TeamID); err == nil {
			teamID = &id
		}
	}
	
	empID := uuid.New()

	var doj, dob *time.Time
	if req.DateOfJoining != "" {
		if t, err := time.Parse(time.RFC3339, req.DateOfJoining); err == nil {
			doj = &t
		} else if t, err := time.Parse(time.DateOnly, req.DateOfJoining); err == nil {
			doj = &t
		}
	}
	if req.DateOfBirth != "" {
		if t, err := time.Parse(time.RFC3339, req.DateOfBirth); err == nil {
			dob = &t
		} else if t, err := time.Parse(time.DateOnly, req.DateOfBirth); err == nil {
			dob = &t
		}
	}
	var managerID *uuid.UUID
	if req.ReportingManager != "" {
		if id, err := uuid.Parse(req.ReportingManager); err == nil {
			managerID = &id
		}
	}

	newUser := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		UserType: 2, // Employee user type
	}

	newEmployee := models.Employee{
		ID:                     empID,
		OrganizationID:         orgID,
		EmployeeCode:           req.EmployeeCode,
		Name:                   req.Name,
		Email:                  req.Email,
		PasswordHash:           string(hashedPassword),
		Phone:                  req.Phone,
		DateOfJoining:          doj,
		Designation:            req.Designation,
		DepartmentID:           deptID,
		RoleID:                 roleID,
		TeamID:                 teamID,
		ReportingManager:       managerID,
		WorkLocation:           req.WorkLocation,
		EmploymentType:         req.EmploymentType,
		Photo:                  req.Photo,
		Address:                req.Address,
		DateOfBirth:            dob,
		Gender:                 req.Gender,
		BloodGroup:             req.BloodGroup,
		EmergencyContactName:   req.EmergencyContactName,
		EmergencyContactNumber: req.EmergencyContactNumber,
		Relationship:           req.Relationship,
		Notes:                  req.Notes,
		Status:                 status,
	}

	err = h.Repo.CreateWithUser(&newEmployee, &newUser)

	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send email in background
	go func(email, name, code string) {
		utils.LogInfo("Attempting to send welcome email to %s...", email)
		subject := "Welcome to Handdy - Your Account Details"
		body := fmt.Sprintf("Hello %s,\n\nYour account has been created successfully.\nYour login credentials are:\nEmail: %s\nEmployee Code: %s\n\nPlease change your password after logging in.\n\nThanks,\nHanddy Team", name, email, code)
		err := h.EmailService.SendEmail([]string{email}, subject, body)
		if err != nil {
			utils.LogError("Failed to send welcome email to %s: %v", email, err)
		} else {
			utils.LogSuccess("Welcome email sent successfully to %s", email)
		}
	}(req.Email, req.Name, req.EmployeeCode)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"message":  "Employee created successfully",
		"employee": newEmployee,
	})
}

// GET /api/v1/employees
func (h *EmployeeHandler) List(w http.ResponseWriter, r *http.Request) {
	page, limit, offset := utils.GetPaginationParams(r)

	employees, totalCount, err := h.Repo.FindAll(limit, offset)
	if err != nil {
		jsonError(w, "Failed to fetch employees", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(utils.PaginatedResponse{
		Success:    true,
		Data:       employees,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: utils.CalculateTotalPages(totalCount, limit),
	})
}

// GET /api/v1/employees/{id}
func (h *EmployeeHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		jsonError(w, "Invalid employee ID", http.StatusBadRequest)
		return
	}

	employee, err := h.Repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			jsonError(w, "Employee not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"employee": employee,
	})
}

// PUT /api/v1/employees/{id}
func (h *EmployeeHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		jsonError(w, "Invalid employee ID", http.StatusBadRequest)
		return
	}

	employee, err := h.Repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			jsonError(w, "Employee not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}

	var req UpdateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update fields if provided
	if name := strings.TrimSpace(req.Name); name != "" {
		employee.Name = name
	}
	if email := strings.ToLower(strings.TrimSpace(req.Email)); email != "" {
		employee.Email = email
	}
	if req.EmployeeCode != "" {
		employee.EmployeeCode = req.EmployeeCode
	}
	if req.Phone != "" {
		employee.Phone = req.Phone
	}
	if req.DateOfJoining != "" {
		if t, err := time.Parse(time.RFC3339, req.DateOfJoining); err == nil {
			employee.DateOfJoining = &t
		} else if t, err := time.Parse(time.DateOnly, req.DateOfJoining); err == nil {
			employee.DateOfJoining = &t
		}
	}
	if req.Designation != "" {
		employee.Designation = req.Designation
	}
	if req.Status != "" {
		employee.Status = req.Status
	}
	if req.WorkLocation != "" {
		employee.WorkLocation = req.WorkLocation
	}
	if req.EmploymentType != "" {
		employee.EmploymentType = req.EmploymentType
	}
	if req.Photo != "" {
		employee.Photo = req.Photo
	}
	if req.Address != "" {
		employee.Address = req.Address
	}
	if req.DateOfBirth != "" {
		if t, err := time.Parse(time.RFC3339, req.DateOfBirth); err == nil {
			employee.DateOfBirth = &t
		} else if t, err := time.Parse(time.DateOnly, req.DateOfBirth); err == nil {
			employee.DateOfBirth = &t
		}
	}
	if req.Gender != "" {
		employee.Gender = req.Gender
	}
	if req.BloodGroup != "" {
		employee.BloodGroup = req.BloodGroup
	}
	if req.EmergencyContactName != "" {
		employee.EmergencyContactName = req.EmergencyContactName
	}
	if req.EmergencyContactNumber != "" {
		employee.EmergencyContactNumber = req.EmergencyContactNumber
	}
	if req.Relationship != "" {
		employee.Relationship = req.Relationship
	}
	if req.Notes != "" {
		employee.Notes = req.Notes
	}

	if orgID, err := uuid.Parse(req.OrganizationID); err == nil && req.OrganizationID != "" {
		employee.OrganizationID = orgID
	}
	if deptID, err := uuid.Parse(req.DepartmentID); err == nil && req.DepartmentID != "" {
		employee.DepartmentID = deptID
	}
	if roleID, err := uuid.Parse(req.RoleID); err == nil && req.RoleID != "" {
		employee.RoleID = roleID
	}
	if tID, err := uuid.Parse(req.TeamID); err == nil && req.TeamID != "" {
		employee.TeamID = &tID
	}
	if managerID, err := uuid.Parse(req.ReportingManager); err == nil && req.ReportingManager != "" {
		employee.ReportingManager = &managerID
	}

	// Update employee and also sync name/email to the users table
	err = h.Repo.UpdateWithUser(employee)

	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"message":  "Employee updated successfully",
		"employee": employee,
	})
}

// DELETE /api/v1/employees/{id}
func (h *EmployeeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		jsonError(w, "Invalid employee ID", http.StatusBadRequest)
		return
	}

	// We don't need to find it first, just call DeleteWithUser
	err = h.Repo.DeleteWithUser(id)

	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Employee deleted successfully",
	})
}

// POST /api/v1/test-email
func (h *EmployeeHandler) TestEmail(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		jsonError(w, "Email is required", http.StatusBadRequest)
		return
	}

	utils.LogInfo("Testing email send to: %s", req.Email)

	err := h.EmailService.SendEmail([]string{req.Email}, "Test Email from Handdy API", "This is a test email to verify that the SMTP configuration is working correctly.")
	if err != nil {
		utils.LogError("Test email failed: %v", err)
		jsonError(w, "Failed to send test email: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Test email sent successfully",
	})
}
