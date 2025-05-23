package handlers

import (
	"encoding/json"
	"fmt"
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

// InitializeLeagueHandler handles POST /api/leagues/initialize
func (lh *LeagueHandler) InitializeLeagueHandler(w http.ResponseWriter, r *http.Request) {
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

	// Start transaction-like behavior with multiple operations
	ctx := r.Context()

	// 1. Create the league
	league, err := lh.db.CreateLeague(ctx, &req)
	if err != nil {
		log.Printf("Failed to create league: %v", err)
		http.Error(w, "Failed to create league", http.StatusInternalServerError)
		return
	}

	// 2. Get default teams
	teams, err := lh.db.GetDefaultTeams(ctx)
	if err != nil {
		log.Printf("Failed to get default teams: %v", err)
		http.Error(w, "Failed to get default teams", http.StatusInternalServerError)
		return
	}

	// 3. Add teams to league and initialize standings
	for _, team := range teams {
		// Add team to league
		if err := lh.db.AddTeamToLeague(ctx, league.ID, team.ID); err != nil {
			log.Printf("Failed to add team %d to league %d: %v", team.ID, league.ID, err)
			http.Error(w, "Failed to add teams to league", http.StatusInternalServerError)
			return
		}

		// Initialize standings for the team
		if err := lh.db.InitializeStanding(ctx, league.ID, team.ID); err != nil {
			log.Printf("Failed to initialize standing for team %d in league %d: %v", team.ID, league.ID, err)
			http.Error(w, "Failed to initialize standings", http.StatusInternalServerError)
			return
		}
	}

	// Convert teams to response format
	var teamResponses []models.Team
	for _, team := range teams {
		teamResponses = append(teamResponses, *team)
	}

	// Create response
	resp := models.InitializeLeagueResponse{
		League: models.LeagueResponse{
			ID:          league.ID,
			Name:        league.Name,
			Status:      league.Status,
			CurrentWeek: league.CurrentWeek,
			CreatedAt:   league.CreatedAt,
		},
		Teams:   teamResponses,
		Message: fmt.Sprintf("League '%s' initialized successfully with %d teams", league.Name, len(teams)),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}
