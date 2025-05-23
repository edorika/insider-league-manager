package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
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

// AddTeamToLeagueHandler handles POST /api/leagues/add-team/:leagueID/:teamID
func (lh *LeagueHandler) AddTeamToLeagueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract leagueID and teamID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) != 5 || pathParts[0] != "api" || pathParts[1] != "leagues" || pathParts[2] != "add-team" {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}

	leagueID, err := strconv.Atoi(pathParts[3])
	if err != nil {
		http.Error(w, "Invalid league ID", http.StatusBadRequest)
		return
	}

	teamID, err := strconv.Atoi(pathParts[4])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// 1. Validate league exists
	league, err := lh.db.GetLeagueByID(ctx, leagueID)
	if err != nil {
		log.Printf("Failed to get league by ID %d: %v", leagueID, err)
		if strings.Contains(err.Error(), "no rows") {
			http.Error(w, "League not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get league", http.StatusInternalServerError)
		}
		return
	}

	// 2. Validate team exists
	team, err := lh.db.GetTeamByID(ctx, teamID)
	if err != nil {
		log.Printf("Failed to get team by ID %d: %v", teamID, err)
		if strings.Contains(err.Error(), "no rows") {
			http.Error(w, "Team not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get team", http.StatusInternalServerError)
		}
		return
	}

	// 3. Add team to league
	if err := lh.db.AddTeamToLeague(ctx, leagueID, teamID); err != nil {
		log.Printf("Failed to add team %d to league %d: %v", teamID, leagueID, err)
		http.Error(w, "Failed to add team to league", http.StatusInternalServerError)
		return
	}

	// 4. Initialize standings for the team
	if err := lh.db.InitializeStanding(ctx, leagueID, teamID); err != nil {
		log.Printf("Failed to initialize standing for team %d in league %d: %v", teamID, leagueID, err)
		http.Error(w, "Failed to initialize standings", http.StatusInternalServerError)
		return
	}

	// Create response
	resp := models.AddTeamToLeagueResponse{
		League: models.LeagueResponse{
			ID:          league.ID,
			Name:        league.Name,
			Status:      league.Status,
			CurrentWeek: league.CurrentWeek,
			CreatedAt:   league.CreatedAt,
		},
		Team: models.Team{
			ID:       team.ID,
			Name:     team.Name,
			Strength: team.Strength,
		},
		Message: fmt.Sprintf("Team '%s' added to league '%s' successfully", team.Name, league.Name),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// RemoveTeamFromLeagueHandler handles POST /api/leagues/remove-team/:leagueID/:teamID
func (lh *LeagueHandler) RemoveTeamFromLeagueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract leagueID and teamID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) != 5 || pathParts[0] != "api" || pathParts[1] != "leagues" || pathParts[2] != "remove-team" {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}

	leagueID, err := strconv.Atoi(pathParts[3])
	if err != nil {
		http.Error(w, "Invalid league ID", http.StatusBadRequest)
		return
	}

	teamID, err := strconv.Atoi(pathParts[4])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// 1. Validate league exists
	league, err := lh.db.GetLeagueByID(ctx, leagueID)
	if err != nil {
		log.Printf("Failed to get league by ID %d: %v", leagueID, err)
		if strings.Contains(err.Error(), "no rows") {
			http.Error(w, "League not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get league", http.StatusInternalServerError)
		}
		return
	}

	// 2. Validate team exists
	team, err := lh.db.GetTeamByID(ctx, teamID)
	if err != nil {
		log.Printf("Failed to get team by ID %d: %v", teamID, err)
		if strings.Contains(err.Error(), "no rows") {
			http.Error(w, "Team not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get team", http.StatusInternalServerError)
		}
		return
	}

	// 3. Remove team from league
	if err := lh.db.RemoveTeamFromLeague(ctx, leagueID, teamID); err != nil {
		log.Printf("Failed to remove team %d from league %d: %v", teamID, leagueID, err)
		if strings.Contains(err.Error(), "is not in league") {
			http.Error(w, "Team is not in this league", http.StatusBadRequest)
		} else {
			http.Error(w, "Failed to remove team from league", http.StatusInternalServerError)
		}
		return
	}

	// Create response
	resp := models.RemoveTeamFromLeagueResponse{
		League: models.LeagueResponse{
			ID:          league.ID,
			Name:        league.Name,
			Status:      league.Status,
			CurrentWeek: league.CurrentWeek,
			CreatedAt:   league.CreatedAt,
		},
		Team: models.Team{
			ID:       team.ID,
			Name:     team.Name,
			Strength: team.Strength,
		},
		Message: fmt.Sprintf("Team '%s' removed from league '%s' successfully", team.Name, league.Name),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// StartLeagueHandler handles POST /api/leagues/start/:leagueID
func (lh *LeagueHandler) StartLeagueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract leagueID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) != 4 || pathParts[0] != "api" || pathParts[1] != "leagues" || pathParts[2] != "start" {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}

	leagueID, err := strconv.Atoi(pathParts[3])
	if err != nil {
		http.Error(w, "Invalid league ID", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// 1. Validate league exists and get its current state
	league, err := lh.db.GetLeagueByID(ctx, leagueID)
	if err != nil {
		log.Printf("Failed to get league by ID %d: %v", leagueID, err)
		if strings.Contains(err.Error(), "no rows") {
			http.Error(w, "League not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get league", http.StatusInternalServerError)
		}
		return
	}

	// 2. Check if league is in correct status to start
	if league.Status != "created" {
		http.Error(w, fmt.Sprintf("League is already %s. Only 'created' leagues can be started", league.Status), http.StatusBadRequest)
		return
	}

	// 3. Get all teams in the league
	teams, err := lh.db.GetTeamsInLeague(ctx, leagueID)
	if err != nil {
		log.Printf("Failed to get teams in league %d: %v", leagueID, err)
		http.Error(w, "Failed to get teams in league", http.StatusInternalServerError)
		return
	}

	// 4. Validate minimum teams (need at least 2 teams to make matches)
	if len(teams) < 2 {
		http.Error(w, "League must have at least 2 teams to start", http.StatusBadRequest)
		return
	}

	// 5. Generate round-robin match schedule
	matches := lh.generateRoundRobinMatches(teams, leagueID)

	// 6. Create all matches in database
	createdMatches := 0
	for _, match := range matches {
		_, err := lh.db.CreateMatch(ctx, &match)
		if err != nil {
			log.Printf("Failed to create match: %v", err)
			http.Error(w, "Failed to create match schedule", http.StatusInternalServerError)
			return
		}
		createdMatches++
	}

	// 7. Update league status to "started"
	if err := lh.db.UpdateLeagueStatus(ctx, leagueID, "started"); err != nil {
		log.Printf("Failed to update league status: %v", err)
		http.Error(w, "Failed to update league status", http.StatusInternalServerError)
		return
	}

	// 8. Calculate total weeks
	totalWeeks := lh.calculateTotalWeeks(len(teams))

	// Create response
	resp := models.StartLeagueResponse{
		League: models.LeagueResponse{
			ID:          league.ID,
			Name:        league.Name,
			Status:      "started",
			CurrentWeek: league.CurrentWeek,
			CreatedAt:   league.CreatedAt,
		},
		TeamsCount:   len(teams),
		MatchesCount: createdMatches,
		TotalWeeks:   totalWeeks,
		Message:      fmt.Sprintf("League '%s' started successfully with %d teams and %d matches scheduled over %d weeks", league.Name, len(teams), createdMatches, totalWeeks),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// generateRoundRobinMatches creates a Premier League style schedule where each team plays every other team twice (home and away)
// First half: each team plays every other team once
// Second half: each team plays every other team again with home/away reversed
func (lh *LeagueHandler) generateRoundRobinMatches(teams []*models.Team, leagueID int) []models.Match {
	var matches []models.Match
	n := len(teams)

	if n < 2 {
		return matches
	}

	week := 1

	// First half of the season: each team plays every other team once
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			// Create match: team i at home vs team j away
			match := models.Match{
				LeagueID:   leagueID,
				HomeTeamID: teams[i].ID,
				AwayTeamID: teams[j].ID,
				Week:       week,
				Status:     "scheduled",
			}
			matches = append(matches, match)

			// Move to next week after scheduling matches
			// For even number of teams, we can have n/2 matches per week
			// For odd number of teams, we can have (n-1)/2 matches per week
			matchesPerWeek := n / 2
			if n%2 == 1 {
				matchesPerWeek = (n - 1) / 2
			}

			// Check if we should move to next week
			if len(matches)%matchesPerWeek == 0 {
				week++
			}
		}
	}

	// Calculate weeks for first half
	firstHalfWeeks := week - 1

	// Second half of the season: reverse home/away for each first half match

	// Create second half matches by reversing home/away from first half
	firstHalfMatches := make([]models.Match, len(matches))
	copy(firstHalfMatches, matches)

	for _, firstHalfMatch := range firstHalfMatches {
		// Create reverse fixture: away team becomes home, home team becomes away
		reverseMatch := models.Match{
			LeagueID:   leagueID,
			HomeTeamID: firstHalfMatch.AwayTeamID,            // Swap home and away
			AwayTeamID: firstHalfMatch.HomeTeamID,            // Swap home and away
			Week:       firstHalfMatch.Week + firstHalfWeeks, // Add to second half
			Status:     "scheduled",
		}
		matches = append(matches, reverseMatch)
	}

	return matches
}

// calculateTotalWeeks calculates the total number of weeks needed for the league (including both halves)
func (lh *LeagueHandler) calculateTotalWeeks(numTeams int) int {
	if numTeams < 2 {
		return 0
	}

	// Each team plays every other team twice (home and away)
	// First half: (n-1) weeks, Second half: (n-1) weeks
	// Total: 2 * (n-1) weeks
	return 2 * (numTeams - 1)
}
