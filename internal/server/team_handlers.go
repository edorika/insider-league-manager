package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"insider-league-manager/internal/models"
)

// createTeamHandler handles POST /api/teams
func (s *Server) createTeamHandler(w http.ResponseWriter, r *http.Request) {
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
	team, err := s.db.CreateTeam(r.Context(), &req)
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
