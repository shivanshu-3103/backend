package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"go-api/models"
	"go-api/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB           *gorm.DB
	EmployeeRepo repository.EmployeeRepository
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type EmployeeLoginRequest struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	DeviceFingerprint string `json:"device_fingerprint"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {

	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if req.Name == "" {
		jsonError(w, "Name is required", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		jsonError(w, "Email is required", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 6 {
		jsonError(w, "Password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	// Check existing user
	var existingUser models.User
	err := h.DB.Where("email = ?", req.Email).First(&existingUser).Error

	if err == nil {
		jsonError(w, "Email already registered", http.StatusConflict)
		return
	}

	if err != gorm.ErrRecordNotFound {
		jsonError(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		jsonError(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	newUser := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := h.DB.Create(&newUser).Error; err != nil {
		jsonError(w, fmt.Sprintf("Failed to create user: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User registered successfully",
		"user":    newUser,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {

	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	var user models.User

	err := h.DB.Where("email = ?", req.Email).First(&user).Error

	if err == gorm.ErrRecordNotFound {
		jsonError(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	if err != nil {
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}

	if user.IsBanned {
		jsonError(w, "User is banned", http.StatusForbidden)
		return
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(req.Password),
	); err != nil {
		jsonError(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	token, err := generateToken(user.ID, user.Email)

	if err != nil {
		jsonError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Login successful",
		"token":   token,
		"user":    user,
	}

	// If user is an employee, fetch employee context
	if user.UserType == 2 && h.EmployeeRepo != nil {
		var employee models.Employee
		// We have to find employee by email since we don't store employee_id in users
		if err := h.DB.Where("email = ?", user.Email).First(&employee).Error; err == nil {
			response["employee_id"] = employee.ID
			response["employee"] = employee
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) EmployeeLogin(w http.ResponseWriter, r *http.Request) {

	var req EmployeeLoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	var user models.User

	err := h.DB.Where("email = ?", req.Email).First(&user).Error

	if err == gorm.ErrRecordNotFound {
		jsonError(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	if err != nil {
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}

	if user.IsBanned {
		jsonError(w, "User is banned", http.StatusForbidden)
		return
	}

	if user.UserType != 2 {
		jsonError(w, "Only employees can login through this endpoint", http.StatusForbidden)
		return
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(req.Password),
	); err != nil {
		jsonError(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	token, err := generateToken(user.ID, user.Email)

	if err != nil {
		jsonError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Login successful",
		"token":   token,
		"user":    user,
	}

	// Fetch employee context
	var employee models.Employee
	if err := h.DB.Where("email = ?", user.Email).First(&employee).Error; err == nil {
		response["employee_id"] = employee.ID
		response["employee"] = employee
		
		isDeviceRegistered := false
		if req.DeviceFingerprint != "" {
			var device models.EmployeeDevice
			err := h.DB.Where("employee_id = ? AND device_fingerprint = ?", employee.ID, req.DeviceFingerprint).First(&device).Error
			if err == nil {
				isDeviceRegistered = true
			}
		}
		response["is_device_registered"] = isDeviceRegistered
	} else {
		response["is_device_registered"] = false
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func generateToken(userID int, email string) (string, error) {

	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString(
		[]byte(os.Getenv("JWT_SECRET")),
	)
}
