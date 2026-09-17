package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	"go-api/middleware"
	"go-api/models"
	"go-api/repository"
	"go-api/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TeamHandler struct {
	Repo repository.TeamRepository
	DB   *gorm.DB
}

func (h *TeamHandler) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	var user models.User
	if err := h.DB.First(&user, middleware.GetUserID(r)).Error; err != nil || user.UserType == 2 {
		jsonError(w, "Admin access required", http.StatusForbidden)
		return false
	}
	return true
}

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	return slug
}

type SaveTeamRequest struct {
	Name string `json:"name"`
}

func (h *TeamHandler) AdminCreateTeam(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}

	var req SaveTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		jsonError(w, "Team name is required", http.StatusBadRequest)
		return
	}

	slug := generateSlug(req.Name)
	team := &models.Team{
		Name: req.Name,
		Slug: slug,
	}

	if err := h.Repo.Create(team); err != nil {
		if strings.Contains(err.Error(), "unique") {
			jsonError(w, "Team with this slug already exists", http.StatusConflict)
		} else {
			jsonError(w, "Failed to create team", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "team": team})
}

func (h *TeamHandler) AdminListTeams(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	page, limit, offset := utils.GetPaginationParams(r)

	teams, totalCount, err := h.Repo.FindAll(limit, offset)
	if err != nil {
		jsonError(w, "Failed to fetch teams", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(utils.PaginatedResponse{
		Success:    true,
		Data:       teams,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: utils.CalculateTotalPages(totalCount, limit),
	})
}

func (h *TeamHandler) AdminUpdateTeam(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		jsonError(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	var req SaveTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	team, err := h.Repo.FindByID(id)
	if err != nil {
		jsonError(w, "Team not found", http.StatusNotFound)
		return
	}

	if req.Name != "" {
		team.Name = req.Name
		team.Slug = generateSlug(req.Name)
	}

	if err := h.Repo.Update(team); err != nil {
		if strings.Contains(err.Error(), "unique") {
			jsonError(w, "Team with this slug already exists", http.StatusConflict)
		} else {
			jsonError(w, "Failed to update team", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "team": team})
}

func (h *TeamHandler) AdminDeleteTeam(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		jsonError(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	if err := h.Repo.Delete(id); err != nil {
		jsonError(w, "Failed to delete team", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Team deleted successfully"})
}
