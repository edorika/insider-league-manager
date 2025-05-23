package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"insider-league-manager/internal/database"
	"insider-league-manager/internal/models"
)

type LeagueHandler struct {
	db database.Service
}

func NewLeagueHandler(db database.Service) *LeagueHandler {
	return &LeagueHandler{
		db: db,
	}
}

// CreateLeagueHandler handles POST /api/leagues/create
func (lh *LeagueHandler) CreateLeagueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateLeagueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Basic validation
	if strings.TrimSpace(req.Name) == "" {
		http.Error(w, "League name is required", http.StatusBadRequest)
		return
	}

	// Create the league
	league, err := lh.db.CreateLeague(r.Context(), &req)
	if err != nil {
		log.Printf("Failed to create league: %v", err)
		http.Error(w, "Failed to create league", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	resp := models.LeagueResponse{
		ID:          league.ID,
		Name:        league.Name,
		Status:      league.Status,
		CurrentWeek: league.CurrentWeek,
		CreatedAt:   league.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}
