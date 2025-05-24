package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

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
// First half: each team plays every other team once, properly distributed across weeks
// Second half: each team plays every other team again with home/away reversed
func (lh *LeagueHandler) generateRoundRobinMatches(teams []*models.Team, leagueID int) []models.Match {
	var matches []models.Match
	n := len(teams)

	if n < 2 {
		return matches
	}

	// For proper round-robin scheduling, we need to handle even and odd number of teams
	if n%2 == 1 {
		// Add a "bye" team for odd number of teams to make scheduling easier
		byeTeam := &models.Team{ID: -1, Name: "BYE"}
		teams = append(teams, byeTeam)
		n = len(teams)
	}

	var firstHalfMatches []models.Match

	// Generate first half using round-robin algorithm
	// Each round has n/2 matches, and we need n-1 rounds for everyone to play everyone once
	for round := 0; round < n-1; round++ {
		weekMatches := lh.generateRoundMatches(teams, round)

		for _, match := range weekMatches {
			// Skip matches involving the "bye" team
			if match.HomeTeamID == -1 || match.AwayTeamID == -1 {
				continue
			}

			match.LeagueID = leagueID
			match.Week = round + 1
			match.Status = "scheduled"
			firstHalfMatches = append(firstHalfMatches, match)
		}
	}

	// Add first half matches to total
	matches = append(matches, firstHalfMatches...)

	// Generate second half by reversing home/away for each first half match
	firstHalfWeeks := n - 1
	for _, firstHalfMatch := range firstHalfMatches {
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

// generateRoundMatches generates matches for a specific round using round-robin algorithm
func (lh *LeagueHandler) generateRoundMatches(teams []*models.Team, round int) []models.Match {
	var matches []models.Match
	n := len(teams)

	// In round-robin, team 0 is fixed, others rotate
	// The algorithm pairs teams in a specific pattern for each round

	for i := 0; i < n/2; i++ {
		var homeTeam, awayTeam *models.Team

		if i == 0 {
			// Team 0 is always fixed
			homeTeam = teams[0]
			// The opponent rotates: in round r, team 0 plays team (r+1)
			awayIndex := (round + 1) % (n - 1)
			if awayIndex == 0 {
				awayIndex = n - 1
			}
			awayTeam = teams[awayIndex]
		} else {
			// For other matches, calculate the pairing
			homeIndex := ((round - i + n - 1) % (n - 1)) + 1
			awayIndex := ((round + i) % (n - 1)) + 1

			homeTeam = teams[homeIndex]
			awayTeam = teams[awayIndex]
		}

		// Alternate home/away advantage across rounds
		if round%2 == 1 && i > 0 {
			homeTeam, awayTeam = awayTeam, homeTeam
		}

		match := models.Match{
			HomeTeamID: homeTeam.ID,
			AwayTeamID: awayTeam.ID,
		}
		matches = append(matches, match)
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

// AdvanceWeekHandler handles POST /api/leagues/advance-week/:leagueID
func (lh *LeagueHandler) AdvanceWeekHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract leagueID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) != 4 || pathParts[0] != "api" || pathParts[1] != "leagues" || pathParts[2] != "advance-week" {
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

	// 2. Check if league is in correct status to advance
	if league.Status != "started" {
		http.Error(w, fmt.Sprintf("League must be 'started' to advance weeks. Current status: %s", league.Status), http.StatusBadRequest)
		return
	}

	// 3. Calculate which week to play (current_week + 1)
	weekToPlay := league.CurrentWeek + 1

	// 4. Get all matches for the week to be played
	matches, err := lh.db.GetMatchesByWeekAndLeague(ctx, leagueID, weekToPlay)
	if err != nil {
		log.Printf("Failed to get matches for league %d week %d: %v", leagueID, weekToPlay, err)
		http.Error(w, "Failed to get matches for the week", http.StatusInternalServerError)
		return
	}

	// 5. If no matches for this week, the league might be finished
	if len(matches) == 0 {
		http.Error(w, "No matches found for the next week. League may be finished.", http.StatusBadRequest)
		return
	}

	// 6. Play all matches for this week
	var matchResults []models.MatchResult
	for _, match := range matches {
		// DEBUG: Log match status before playing
		log.Printf("DEBUG: Playing match ID %d, status: %s, home_goals: %v, away_goals: %v",
			match.ID, match.Status, match.HomeGoals, match.AwayGoals)

		// Generate match result based on team strengths
		homeGoals, awayGoals := lh.generateMatchResult(match.HomeTeamID, match.AwayTeamID)
		log.Printf("DEBUG: Generated result for match %d: %d-%d", match.ID, homeGoals, awayGoals)

		// Update match in database
		if err := lh.db.PlayMatch(ctx, match.ID, homeGoals, awayGoals); err != nil {
			log.Printf("Failed to play match %d: %v", match.ID, err)
			http.Error(w, "Failed to play matches", http.StatusInternalServerError)
			return
		}
		log.Printf("DEBUG: Successfully updated match %d in database with %d-%d", match.ID, homeGoals, awayGoals)

		// Update standings
		if err := lh.db.UpdateStandings(ctx, leagueID, match.HomeTeamID, match.AwayTeamID, homeGoals, awayGoals); err != nil {
			log.Printf("Failed to update standings for match %d: %v", match.ID, err)
			http.Error(w, "Failed to update standings", http.StatusInternalServerError)
			return
		}

		// Get team names for response
		homeTeam, err := lh.db.GetTeamByID(ctx, match.HomeTeamID)
		if err != nil {
			log.Printf("Failed to get home team %d: %v", match.HomeTeamID, err)
			http.Error(w, "Failed to get team information", http.StatusInternalServerError)
			return
		}

		awayTeam, err := lh.db.GetTeamByID(ctx, match.AwayTeamID)
		if err != nil {
			log.Printf("Failed to get away team %d: %v", match.AwayTeamID, err)
			http.Error(w, "Failed to get team information", http.StatusInternalServerError)
			return
		}

		// Update match object with played results for response
		match.HomeGoals = &homeGoals
		match.AwayGoals = &awayGoals
		match.Status = "played"
		log.Printf("DEBUG: Match object updated for response: %d-%d", *match.HomeGoals, *match.AwayGoals)

		// Create match result for response
		matchResult := models.MatchResult{
			Match:    *match,
			HomeTeam: homeTeam.Name,
			AwayTeam: awayTeam.Name,
			Result:   fmt.Sprintf("%d-%d", homeGoals, awayGoals),
		}
		matchResults = append(matchResults, matchResult)
	}

	// 7. Advance the league week
	if err := lh.db.AdvanceLeagueWeek(ctx, leagueID); err != nil {
		log.Printf("Failed to advance league %d week: %v", leagueID, err)
		http.Error(w, "Failed to advance league week", http.StatusInternalServerError)
		return
	}

	// 8. Check if league is finished (no more matches)
	nextWeek := weekToPlay + 1
	nextWeekMatches, err := lh.db.GetMatchesByWeekAndLeague(ctx, leagueID, nextWeek)
	if err != nil {
		log.Printf("Failed to check next week matches: %v", err)
		// Continue anyway, this is not critical
	}

	// If no matches next week, mark league as finished
	if len(nextWeekMatches) == 0 {
		if err := lh.db.UpdateLeagueStatus(ctx, leagueID, "finished"); err != nil {
			log.Printf("Failed to mark league as finished: %v", err)
			// Continue anyway, this is not critical
		}
		league.Status = "finished"
	}

	// Update league current week for response
	league.CurrentWeek = weekToPlay

	// Create response
	resp := models.AdvanceWeekResponse{
		League: models.LeagueResponse{
			ID:          league.ID,
			Name:        league.Name,
			Status:      league.Status,
			CurrentWeek: league.CurrentWeek,
			CreatedAt:   league.CreatedAt,
		},
		WeekAdvanced:  weekToPlay,
		MatchesPlayed: matchResults,
		Message:       fmt.Sprintf("League '%s' advanced to week %d. %d matches played.", league.Name, weekToPlay, len(matchResults)),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// generateMatchResult simulates a football match using team strengths to influence the result
func (lh *LeagueHandler) generateMatchResult(homeTeamID, awayTeamID int) (int, int) {
	// Get team strengths (we already validated teams exist earlier in the flow)
	homeTeam, err := lh.db.GetTeamByID(context.Background(), homeTeamID)
	if err != nil {
		// Fallback to basic random if we can't get team info
		return lh.basicRandomGoals(), lh.basicRandomGoals()
	}

	awayTeam, err := lh.db.GetTeamByID(context.Background(), awayTeamID)
	if err != nil {
		// Fallback to basic random if we can't get team info
		return lh.basicRandomGoals(), lh.basicRandomGoals()
	}

	// Simulate match based on team strengths
	log.Printf("DEBUG: Team strengths - Home: %s (%d), Away: %s (%d)", homeTeam.Name, homeTeam.Strength, awayTeam.Name, awayTeam.Strength)
	return lh.simulateMatch(homeTeam.Strength, awayTeam.Strength)
}

// simulateMatch generates realistic match results based on team strengths
func (lh *LeagueHandler) simulateMatch(homeStrength, awayStrength int) (int, int) {
	// Add home advantage (typically 3-5 points)
	homeAdvantage := 4
	adjustedHomeStrength := homeStrength + homeAdvantage

	// Calculate strength difference (-100 to +100 range)
	strengthDiff := adjustedHomeStrength - awayStrength

	// Generate base goal expectancy based on strength (1.0 to 3.0 goals per team on average)
	homeGoalExpectancy := 1.5 + float64(strengthDiff)/100.0 // Stronger teams score more
	awayGoalExpectancy := 1.5 - float64(strengthDiff)/100.0 // Weaker teams score less

	// Ensure reasonable bounds (0.5 to 3.0 goals expectancy)
	if homeGoalExpectancy < 0.5 {
		homeGoalExpectancy = 0.5
	}
	if homeGoalExpectancy > 3.0 {
		homeGoalExpectancy = 3.0
	}
	if awayGoalExpectancy < 0.5 {
		awayGoalExpectancy = 0.5
	}
	if awayGoalExpectancy > 3.0 {
		awayGoalExpectancy = 3.0
	}

	// Debug expectancy calculations
	log.Printf("DEBUG: Expectancy - Home: %.2f, Away: %.2f (strengthDiff: %d)", homeGoalExpectancy, awayGoalExpectancy, strengthDiff)

	// Use Poisson-like distribution for goal generation
	homeGoals := lh.generateGoalsFromExpectancy(homeGoalExpectancy)
	awayGoals := lh.generateGoalsFromExpectancy(awayGoalExpectancy)

	log.Printf("DEBUG: Final goals - Home: %d, Away: %d", homeGoals, awayGoals)
	return homeGoals, awayGoals
}

// generateGoalsFromExpectancy generates goals using weighted probability based on expectancy
func (lh *LeagueHandler) generateGoalsFromExpectancy(expectancy float64) int {
	// Use time-based seed with microseconds for better randomness
	rand.Seed(time.Now().UnixNano())

	// Generate a random number 0-99 for easier probability calculation
	randNum := rand.Intn(100)

	// Debug the inputs and random number
	log.Printf("DEBUG: generateGoalsFromExpectancy called with expectancy=%.2f, randNum=%d", expectancy, randNum)

	var goals int

	// Simpler probability distribution based on expectancy
	if expectancy <= 1.0 {
		// Low scoring team: mostly 0-1 goals
		if randNum < 50 {
			goals = 0
		} else if randNum < 85 {
			goals = 1
		} else if randNum < 95 {
			goals = 2
		} else {
			goals = 3
		}
	} else if expectancy <= 2.0 {
		// Medium scoring team: balanced scoring
		if randNum < 25 {
			goals = 0
		} else if randNum < 50 {
			goals = 1
		} else if randNum < 75 {
			goals = 2
		} else if randNum < 90 {
			goals = 3
		} else if randNum < 97 {
			goals = 4
		} else {
			goals = 5
		}
	} else {
		// High scoring team: more goals likely
		if randNum < 15 {
			goals = 0
		} else if randNum < 30 {
			goals = 1
		} else if randNum < 50 {
			goals = 2
		} else if randNum < 70 {
			goals = 3
		} else if randNum < 85 {
			goals = 4
		} else if randNum < 95 {
			goals = 5
		} else {
			goals = 6
		}
	}

	log.Printf("DEBUG: generateGoalsFromExpectancy returning %d goals", goals)
	return goals
}

// ViewMatchesHandler handles GET /api/leagues/view-matches/:leagueID
func (lh *LeagueHandler) ViewMatchesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract leagueID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) != 4 || pathParts[0] != "api" || pathParts[1] != "leagues" || pathParts[2] != "view-matches" {
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

	// 2. Check if league has been started
	if league.Status == "created" {
		http.Error(w, "League has not been started yet. No matches to view.", http.StatusBadRequest)
		return
	}

	// 3. Get all matches for the current week
	matches, err := lh.db.GetMatchesByWeekAndLeague(ctx, leagueID, league.CurrentWeek)
	if err != nil {
		log.Printf("Failed to get matches for league %d week %d: %v", leagueID, league.CurrentWeek, err)
		http.Error(w, "Failed to get matches for the current week", http.StatusInternalServerError)
		return
	}

	// 4. If no matches for current week, return empty result
	if len(matches) == 0 {
		resp := models.ViewMatchesResponse{
			League: models.LeagueResponse{
				ID:          league.ID,
				Name:        league.Name,
				Status:      league.Status,
				CurrentWeek: league.CurrentWeek,
				CreatedAt:   league.CreatedAt,
			},
			CurrentWeek: league.CurrentWeek,
			Matches:     []models.MatchResult{},
			Message:     fmt.Sprintf("No matches found for week %d in league '%s'", league.CurrentWeek, league.Name),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("Failed to encode response: %v", err)
		}
		return
	}

	// 5. Build match results with team information
	var matchResults []models.MatchResult
	for _, match := range matches {
		// Get team names for response
		homeTeam, err := lh.db.GetTeamByID(ctx, match.HomeTeamID)
		if err != nil {
			log.Printf("Failed to get home team %d: %v", match.HomeTeamID, err)
			http.Error(w, "Failed to get team information", http.StatusInternalServerError)
			return
		}

		awayTeam, err := lh.db.GetTeamByID(ctx, match.AwayTeamID)
		if err != nil {
			log.Printf("Failed to get away team %d: %v", match.AwayTeamID, err)
			http.Error(w, "Failed to get team information", http.StatusInternalServerError)
			return
		}

		// Create result string based on match status
		var result string
		if match.Status == "played" && match.HomeGoals != nil && match.AwayGoals != nil {
			result = fmt.Sprintf("%d-%d", *match.HomeGoals, *match.AwayGoals)
		} else {
			result = "Not played yet"
		}

		// Create match result for response
		matchResult := models.MatchResult{
			Match:    *match,
			HomeTeam: homeTeam.Name,
			AwayTeam: awayTeam.Name,
			Result:   result,
		}
		matchResults = append(matchResults, matchResult)
	}

	// 6. Create response
	resp := models.ViewMatchesResponse{
		League: models.LeagueResponse{
			ID:          league.ID,
			Name:        league.Name,
			Status:      league.Status,
			CurrentWeek: league.CurrentWeek,
			CreatedAt:   league.CreatedAt,
		},
		CurrentWeek: league.CurrentWeek,
		Matches:     matchResults,
		Message:     fmt.Sprintf("Matches for week %d in league '%s'", league.CurrentWeek, league.Name),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// PlayAllMatchesHandler handles POST /api/leagues/play-all-matches/:leagueID
func (lh *LeagueHandler) PlayAllMatchesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract leagueID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) != 4 || pathParts[0] != "api" || pathParts[1] != "leagues" || pathParts[2] != "play-all-matches" {
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

	// 2. Check if league is in correct status to play matches
	if league.Status != "started" {
		http.Error(w, fmt.Sprintf("League must be 'started' to play matches. Current status: %s", league.Status), http.StatusBadRequest)
		return
	}

	// 3. Calculate total weeks for this league
	teams, err := lh.db.GetTeamsInLeague(ctx, leagueID)
	if err != nil {
		log.Printf("Failed to get teams in league %d: %v", leagueID, err)
		http.Error(w, "Failed to get teams in league", http.StatusInternalServerError)
		return
	}

	totalWeeks := lh.calculateTotalWeeks(len(teams))
	startingWeek := league.CurrentWeek
	var allMatchResults []models.WeekResult
	weeksPlayed := 0

	// 4. Play all remaining weeks
	for currentWeek := league.CurrentWeek + 1; currentWeek <= totalWeeks; currentWeek++ {
		// Get all matches for this week
		matches, err := lh.db.GetMatchesByWeekAndLeague(ctx, leagueID, currentWeek)
		if err != nil {
			log.Printf("Failed to get matches for league %d week %d: %v", leagueID, currentWeek, err)
			http.Error(w, "Failed to get matches for week", http.StatusInternalServerError)
			return
		}

		// If no matches for this week, we're done
		if len(matches) == 0 {
			break
		}

		// Play all matches for this week
		var weekMatchResults []models.MatchResult
		for _, match := range matches {
			// Generate match result based on team strengths
			homeGoals, awayGoals := lh.generateMatchResult(match.HomeTeamID, match.AwayTeamID)
			log.Printf("DEBUG: Generated result for match %d (week %d): %d-%d", match.ID, currentWeek, homeGoals, awayGoals)

			// Update match in database
			if err := lh.db.PlayMatch(ctx, match.ID, homeGoals, awayGoals); err != nil {
				log.Printf("Failed to play match %d: %v", match.ID, err)
				http.Error(w, "Failed to play matches", http.StatusInternalServerError)
				return
			}

			// Update standings
			if err := lh.db.UpdateStandings(ctx, leagueID, match.HomeTeamID, match.AwayTeamID, homeGoals, awayGoals); err != nil {
				log.Printf("Failed to update standings for match %d: %v", match.ID, err)
				http.Error(w, "Failed to update standings", http.StatusInternalServerError)
				return
			}

			// Get team names for response
			homeTeam, err := lh.db.GetTeamByID(ctx, match.HomeTeamID)
			if err != nil {
				log.Printf("Failed to get home team %d: %v", match.HomeTeamID, err)
				http.Error(w, "Failed to get team information", http.StatusInternalServerError)
				return
			}

			awayTeam, err := lh.db.GetTeamByID(ctx, match.AwayTeamID)
			if err != nil {
				log.Printf("Failed to get away team %d: %v", match.AwayTeamID, err)
				http.Error(w, "Failed to get team information", http.StatusInternalServerError)
				return
			}

			// Update match object with played results for response
			match.HomeGoals = &homeGoals
			match.AwayGoals = &awayGoals
			match.Status = "played"

			// Create match result for response
			matchResult := models.MatchResult{
				Match:    *match,
				HomeTeam: homeTeam.Name,
				AwayTeam: awayTeam.Name,
				Result:   fmt.Sprintf("%d-%d", homeGoals, awayGoals),
			}
			weekMatchResults = append(weekMatchResults, matchResult)
		}

		// Add week result to all results
		weekResult := models.WeekResult{
			Week:    currentWeek,
			Matches: weekMatchResults,
		}
		allMatchResults = append(allMatchResults, weekResult)

		// Advance the league week
		if err := lh.db.AdvanceLeagueWeek(ctx, leagueID); err != nil {
			log.Printf("Failed to advance league %d week: %v", leagueID, err)
			http.Error(w, "Failed to advance league week", http.StatusInternalServerError)
			return
		}

		weeksPlayed++
		league.CurrentWeek = currentWeek
	}

	// 5. Mark league as finished
	if err := lh.db.UpdateLeagueStatus(ctx, leagueID, "finished"); err != nil {
		log.Printf("Failed to mark league as finished: %v", err)
		http.Error(w, "Failed to update league status", http.StatusInternalServerError)
		return
	}
	league.Status = "finished"

	// 6. Count total matches played
	totalMatchesPlayed := 0
	for _, weekResult := range allMatchResults {
		totalMatchesPlayed += len(weekResult.Matches)
	}

	// 7. Create response
	resp := models.PlayAllMatchesResponse{
		League: models.LeagueResponse{
			ID:          league.ID,
			Name:        league.Name,
			Status:      league.Status,
			CurrentWeek: league.CurrentWeek,
			CreatedAt:   league.CreatedAt,
		},
		StartingWeek:       startingWeek,
		FinalWeek:          league.CurrentWeek,
		WeeksPlayed:        weeksPlayed,
		TotalMatchesPlayed: totalMatchesPlayed,
		WeekResults:        allMatchResults,
		Message:            fmt.Sprintf("League '%s' completed successfully. Played %d weeks with %d total matches.", league.Name, weeksPlayed, totalMatchesPlayed),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// basicRandomGoals generates basic random goals as fallback
func (lh *LeagueHandler) basicRandomGoals() int {
	rand.Seed(time.Now().UnixNano())
	randInt := rand.Intn(100)

	if randInt < 30 {
		return 0
	} else if randInt < 55 {
		return 1
	} else if randInt < 75 {
		return 2
	} else if randInt < 90 {
		return 3
	} else if randInt < 97 {
		return 4
	} else {
		return 5
	}
}
