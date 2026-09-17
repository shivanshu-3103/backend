package v1

import (
	"encoding/json"
	"net/http"

	"go-api/middleware"
	"go-api/models"
	"go-api/repository"
	"go-api/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskHandler struct {
	Repo           repository.TaskRepository
	AssignmentRepo repository.TaskAssignmentRepository
	DB             *gorm.DB
}

func (h *TaskHandler) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	var user models.User
	if err := h.DB.First(&user, middleware.GetUserID(r)).Error; err != nil || user.UserType == 2 {
		jsonError(w, "Admin access required", http.StatusForbidden)
		return false
	}
	return true
}

type SaveTaskRequest struct {
	Name              string `json:"name"`
	Type              string `json:"type"`
	IdleLogoutEnabled bool   `json:"idle_logout_enabled"`
	IdleLogoutMinutes *int   `json:"idle_logout_minutes"`
	Status            string `json:"status"`
}

func (h *TaskHandler) AdminCreateTask(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}

	var req SaveTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r)

	if req.Name == "" || req.Type == "" {
		jsonError(w, "name and type are required", http.StatusBadRequest)
		return
	}

	if req.Type != "work" && req.Type != "break" && req.Type != "other" {
		jsonError(w, "type must be work, break, or other", http.StatusBadRequest)
		return
	}

	status := "active"
	if req.Status != "" {
		status = req.Status
	}

	task := &models.Task{
		Name:              req.Name,
		Type:              req.Type,
		IdleLogoutEnabled: req.IdleLogoutEnabled,
		IdleLogoutMinutes: req.IdleLogoutMinutes,
		Status:            status,
		CreatedBy:         &userID,
	}

	if err := h.Repo.Create(task); err != nil {
		jsonError(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "task": task})
}

func (h *TaskHandler) AdminListTasks(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	page, limit, offset := utils.GetPaginationParams(r)

	tasks, totalCount, err := h.Repo.FindAll(limit, offset)
	if err != nil {
		jsonError(w, "Failed to fetch tasks", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(utils.PaginatedResponse{
		Success:    true,
		Data:       tasks,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: utils.CalculateTotalPages(totalCount, limit),
	})
}

func (h *TaskHandler) AdminUpdateTask(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		jsonError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var req SaveTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	task, err := h.Repo.FindByID(id)
	if err != nil {
		jsonError(w, "Task not found", http.StatusNotFound)
		return
	}

	if req.Name != "" {
		task.Name = req.Name
	}
	if req.Type != "" {
		if req.Type != "work" && req.Type != "break" && req.Type != "other" {
			jsonError(w, "type must be work, break, or other", http.StatusBadRequest)
			return
		}
		task.Type = req.Type
	}
	task.IdleLogoutEnabled = req.IdleLogoutEnabled
	task.IdleLogoutMinutes = req.IdleLogoutMinutes
	if req.Status != "" {
		task.Status = req.Status
	}

	if err := h.Repo.Update(task); err != nil {
		jsonError(w, "Failed to update task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "task": task})
}

func (h *TaskHandler) AdminDeleteTask(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		jsonError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	if err := h.Repo.Delete(id); err != nil {
		jsonError(w, "Failed to delete task", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Task deleted successfully"})
}

type SaveTaskAssignmentRequest struct {
	AssignType string     `json:"assign_type"` // company, team, employee
	AssignID   *uuid.UUID `json:"assign_id"`   // optional
}

func (h *TaskHandler) AdminAssignTask(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	taskID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		jsonError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var req SaveTaskAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.AssignType == "" {
		jsonError(w, "assign_type is required", http.StatusBadRequest)
		return
	}

	assignment := &models.TaskAssignment{
		TaskID:     taskID,
		AssignType: req.AssignType,
		AssignID:   req.AssignID,
	}

	if err := h.AssignmentRepo.Create(assignment); err != nil {
		jsonError(w, "Failed to create task assignment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "assignment": assignment})
}

func (h *TaskHandler) AdminListTaskAssignments(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	taskID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		jsonError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	page, limit, offset := utils.GetPaginationParams(r)

	assignments, totalCount, err := h.AssignmentRepo.FindByTaskID(taskID, limit, offset)
	if err != nil {
		jsonError(w, "Failed to fetch task assignments", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(utils.PaginatedResponse{
		Success:    true,
		Data:       assignments,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: utils.CalculateTotalPages(totalCount, limit),
	})
}

func (h *TaskHandler) AdminDeleteTaskAssignment(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	assignmentID, err := uuid.Parse(r.PathValue("assignment_id"))
	if err != nil {
		jsonError(w, "Invalid assignment ID", http.StatusBadRequest)
		return
	}

	if err := h.AssignmentRepo.Delete(assignmentID); err != nil {
		jsonError(w, "Failed to delete task assignment", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Assignment deleted successfully"})
}
