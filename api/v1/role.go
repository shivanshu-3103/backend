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

type RoleHandler struct {
	Repo repository.RoleRepository
}

type CreateRoleRequest struct {
	Name         string `json:"name"`
	Status       string `json:"status"`
	DepartmentID string `json:"department_id"`
}

type UpdateRoleRequest struct {
	Name         string `json:"name"`
	Status       string `json:"status"`
	DepartmentID string `json:"department_id"`
}

// POST /api/v1/roles
func (h *RoleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		jsonError(w, "Name is required", http.StatusBadRequest)
		return
	}

	deptID, _ := uuid.Parse(req.DepartmentID)
	status := req.Status
	if status == "" {
		status = "active"
	}

	newRole := models.Role{
		ID:           uuid.New(),
		Name:         req.Name,
		Status:       status,
		DepartmentID: deptID,
	}

	if err := h.Repo.Create(&newRole); err != nil {
		jsonError(w, fmt.Sprintf("Failed to create role: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Role created successfully",
		"role":    newRole,
	})
}

// GET /api/v1/roles
func (h *RoleHandler) List(w http.ResponseWriter, r *http.Request) {
	page, limit, offset := utils.GetPaginationParams(r)

	roles, totalCount, err := h.Repo.FindAll(limit, offset)
	if err != nil {
		jsonError(w, "Failed to fetch roles", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(utils.PaginatedResponse{
		Success:    true,
		Data:       roles,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: utils.CalculateTotalPages(totalCount, limit),
	})
}

// GET /api/v1/departments/{id}/roles
func (h *RoleHandler) ListByDepartment(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	deptID, err := uuid.Parse(idStr)
	if err != nil {
		jsonError(w, "Invalid department ID", http.StatusBadRequest)
		return
	}

	page, limit, offset := utils.GetPaginationParams(r)

	roles, totalCount, err := h.Repo.FindActiveByDepartmentID(deptID, limit, offset)
	if err != nil {
		jsonError(w, "Failed to fetch roles", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(utils.PaginatedResponse{
		Success:    true,
		Data:       roles,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: utils.CalculateTotalPages(totalCount, limit),
	})
}

// GET /api/v1/roles/{id}
func (h *RoleHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		jsonError(w, "Invalid role ID", http.StatusBadRequest)
		return
	}

	role, err := h.Repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			jsonError(w, "Role not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"role":    role,
	})
}

// PUT /api/v1/roles/{id}
func (h *RoleHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		jsonError(w, "Invalid role ID", http.StatusBadRequest)
		return
	}

	role, err := h.Repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			jsonError(w, "Role not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Database error", http.StatusInternalServerError)
		return
	}

	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if name := strings.TrimSpace(req.Name); name != "" {
		role.Name = name
	}
	if req.Status != "" {
		role.Status = req.Status
	}
	if req.DepartmentID != "" {
		if deptID, err := uuid.Parse(req.DepartmentID); err == nil {
			role.DepartmentID = deptID
		}
	}

	if err := h.Repo.Update(role); err != nil {
		jsonError(w, fmt.Sprintf("Failed to update role: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Role updated successfully",
		"role":    role,
	})
}

// DELETE /api/v1/roles/{id}
func (h *RoleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		jsonError(w, "Invalid role ID", http.StatusBadRequest)
		return
	}

	if err := h.Repo.Delete(id); err != nil {
		jsonError(w, "Failed to delete role", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Role deleted successfully",
	})
}
