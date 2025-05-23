package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"insider-league-manager/internal/database"
	"insider-league-manager/internal/models"
)

type TeamHandler struct {
	db database.Service
}

func NewTeamHandler(db database.Service) *TeamHandler {
	return &TeamHandler{
		db: db,
	}
}

// CreateTeamHandler handles POST /api/teams
func (th *TeamHandler) CreateTeamHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Basic validation
	if strings.TrimSpace(req.Name) == "" {
		http.Error(w, "Team name is required", http.StatusBadRequest)
		return
	}

	// Create the team
	team, err := th.db.CreateTeam(r.Context(), &req)
	if err != nil {
		log.Printf("Failed to create team: %v", err)
		http.Error(w, "Failed to create team", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	resp := models.TeamResponse{
		ID:       team.ID,
		Name:     team.Name,
		Strength: team.Strength,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// GetAllTeamsHandler handles GET /api/teams
func (th *TeamHandler) GetAllTeamsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all teams
	teams, err := th.db.GetAllTeams(r.Context())
	if err != nil {
		log.Printf("Failed to get all teams: %v", err)
		http.Error(w, "Failed to get teams", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	var resp []models.TeamResponse
	for _, team := range teams {
		resp = append(resp, models.TeamResponse{
			ID:       team.ID,
			Name:     team.Name,
			Strength: team.Strength,
		})
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// GetTeamByIDHandler handles GET /api/teams/:teamID
func (th *TeamHandler) GetTeamByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract team ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) != 3 || pathParts[0] != "api" || pathParts[1] != "teams" {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}

	teamID, err := strconv.Atoi(pathParts[2])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	// Get team by ID
	team, err := th.db.GetTeamByID(r.Context(), teamID)
	if err != nil {
		log.Printf("Failed to get team by ID %d: %v", teamID, err)
		if strings.Contains(err.Error(), "no rows") {
			http.Error(w, "Team not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get team", http.StatusInternalServerError)
		}
		return
	}

	// Convert to response format
	resp := models.TeamResponse{
		ID:       team.ID,
		Name:     team.Name,
		Strength: team.Strength,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// UpdateTeamHandler handles PUT /api/teams/:teamID
func (th *TeamHandler) UpdateTeamHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract team ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) != 3 || pathParts[0] != "api" || pathParts[1] != "teams" {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}

	teamID, err := strconv.Atoi(pathParts[2])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	var req models.CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Basic validation
	if strings.TrimSpace(req.Name) == "" {
		http.Error(w, "Team name is required", http.StatusBadRequest)
		return
	}

	// Update the team
	team, err := th.db.UpdateTeam(r.Context(), teamID, &req)
	if err != nil {
		log.Printf("Failed to update team with ID %d: %v", teamID, err)
		if strings.Contains(err.Error(), "no rows") {
			http.Error(w, "Team not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to update team", http.StatusInternalServerError)
		}
		return
	}

	// Convert to response format
	resp := models.TeamResponse{
		ID:       team.ID,
		Name:     team.Name,
		Strength: team.Strength,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}
