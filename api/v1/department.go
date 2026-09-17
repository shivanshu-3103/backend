package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go-api/models"
	"go-api/repository"
	"go-api/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DepartmentHandler struct {
	Repo repository.DepartmentRepository
}

type CreateDepartmentRequest struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type UpdateDepartmentRequest struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// POST /api/v1/departments
func (h *DepartmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		jsonError(w, "Name is required", http.StatusBadRequest)
		return
	}

	status := req.Status
	if status == "" {
		status = "active"
	}

	newDept := models.Department{
		ID:     uuid.New(),
		Name:   req.Name,
		Status: status,
	}

	if err := h.Repo.Create(&newDept); err != nil {
		jsonError(w, fmt.Sprintf("Failed to create department: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"message":    "Department created successfully",
		"department": newDept,
	})
}

// GET /api/v1/departments
func (h *DepartmentHandler) List(w http.ResponseWriter, r *http.Request) {
	page, limit, offset := utils.GetPaginationParams(r)

	departments, totalCount, err := h.Repo.FindAll(limit, offset)
	if err != nil {
		jsonError(w, "Failed to fetch departments", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(utils.PaginatedResponse{
		Success:    true,
		Data:       departments,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: utils.CalculateTotalPages(totalCount, limit),
	})
}

// GET /api/v1/departments/{id}
func (h *DepartmentHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		jsonError(w, "Invalid department ID", http.StatusBadRequest)
		return
	}

	department, err := h.Repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			jsonError(w, "Department not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"department": department,
	})
}

// PUT /api/v1/departments/{id}
func (h *DepartmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		jsonError(w, "Invalid department ID", http.StatusBadRequest)
		return
	}

	department, err := h.Repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			jsonError(w, "Department not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}

	var req UpdateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if name := strings.TrimSpace(req.Name); name != "" {
		department.Name = name
	}
	if req.Status != "" {
		department.Status = req.Status
	}

	if err := h.Repo.Update(department); err != nil {
		jsonError(w, fmt.Sprintf("Failed to update department: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"message":    "Department updated successfully",
		"department": department,
	})
}

// DELETE /api/v1/departments/{id}
func (h *DepartmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		jsonError(w, "Invalid department ID", http.StatusBadRequest)
		return
	}

	if err := h.Repo.Delete(id); err != nil {
		jsonError(w, "Failed to delete department", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Department deleted successfully",
	})
}
