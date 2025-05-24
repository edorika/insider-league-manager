package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"insider-league-manager/internal/models"
)

// Mock database service for league testing
type mockLeagueDBService struct {
	*mockDBService // Embed existing mock for team methods
}

func (m *mockLeagueDBService) CreateLeague(ctx context.Context, req *models.CreateLeagueRequest) (*models.League, error) {
	return &models.League{
		ID:          1,
		Name:        req.Name,
		Status:      "created",
		CurrentWeek: 0,
		CreatedAt:   time.Now(),
	}, nil
}

func (m *mockLeagueDBService) AddTeamToLeague(ctx context.Context, leagueID, teamID int) error {
	return nil // Successful operation
}

func (m *mockLeagueDBService) InitializeStanding(ctx context.Context, leagueID, teamID int) error {
	return nil // Successful operation
}

func (m *mockLeagueDBService) GetDefaultTeams(ctx context.Context) ([]*models.Team, error) {
	return []*models.Team{
		{ID: 1, Name: "Manchester City", Strength: 88},
		{ID: 2, Name: "Liverpool FC", Strength: 86},
		{ID: 3, Name: "Chelsea FC", Strength: 84},
		{ID: 4, Name: "Arsenal FC", Strength: 82},
	}, nil
}

func (m *mockLeagueDBService) GetLeagueByID(ctx context.Context, leagueID int) (*models.League, error) {
	switch leagueID {
	case 1:
		// For start league test, return created league; for advance week tests, we'll use league ID 3
		return &models.League{
			ID:          1,
			Name:        "Test League",
			Status:      "created", // Created league for start tests
			CurrentWeek: 0,
			CreatedAt:   time.Now(),
		}, nil
	case 2:
		return &models.League{
			ID:          2,
			Name:        "Created League",
			Status:      "created", // Not started league for testing
			CurrentWeek: 0,
			CreatedAt:   time.Now(),
		}, nil
	case 3:
		return &models.League{
			ID:          3,
			Name:        "Started League",
			Status:      "started", // Started league for advance week tests
			CurrentWeek: 0,
			CreatedAt:   time.Now(),
		}, nil
	default:
		// Return error for any other ID to simulate not found
		return nil, fmt.Errorf("no rows in result set")
	}
}

func (m *mockLeagueDBService) RemoveTeamFromLeague(ctx context.Context, leagueID, teamID int) error {
	// Simulate that team 1 is in league 1, others are not
	if leagueID == 1 && teamID == 1 {
		return nil // Successful removal
	}
	// Return error for any other combination to simulate team not in league
	return fmt.Errorf("team %d is not in league %d", teamID, leagueID)
}

func (m *mockLeagueDBService) GetTeamByID(ctx context.Context, teamID int) (*models.Team, error) {
	if teamID == 1 {
		return &models.Team{
			ID:       1,
			Name:     "Team A",
			Strength: 85,
		}, nil
	}
	if teamID == 2 {
		return &models.Team{
			ID:       2,
			Name:     "Team B",
			Strength: 90,
		}, nil
	}
	// Return error for any other ID to simulate not found
	return nil, fmt.Errorf("no rows in result set")
}

func (m *mockLeagueDBService) GetTeamsInLeague(ctx context.Context, leagueID int) ([]*models.Team, error) {
	if leagueID == 1 {
		return []*models.Team{
			{ID: 1, Name: "Team A", Strength: 85},
			{ID: 2, Name: "Team B", Strength: 90},
		}, nil
	}
	if leagueID == 2 {
		// League with only 1 team (should fail to start)
		return []*models.Team{
			{ID: 1, Name: "Team A", Strength: 85},
		}, nil
	}
	return nil, fmt.Errorf("no teams found in league %d", leagueID)
}

func (m *mockLeagueDBService) CreateMatch(ctx context.Context, match *models.Match) (*models.Match, error) {
	// Return the match with an assigned ID
	createdMatch := *match
	createdMatch.ID = 1
	createdMatch.CreatedAt = time.Now()
	return &createdMatch, nil
}

func (m *mockLeagueDBService) UpdateLeagueStatus(ctx context.Context, leagueID int, status string) error {
	if leagueID == 1 {
		return nil // Successful update
	}
	return fmt.Errorf("no league found with ID %d", leagueID)
}

func (m *mockLeagueDBService) GetMatchesByWeekAndLeague(ctx context.Context, leagueID, week int) ([]*models.Match, error) {
	if leagueID == 3 && week == 1 {
		// Return matches for week 1 for started league (ID 3)
		return []*models.Match{
			{
				ID:         1,
				LeagueID:   3,
				HomeTeamID: 1,
				AwayTeamID: 2,
				Week:       1,
				Status:     "scheduled",
			},
		}, nil
	}
	if leagueID == 3 && week == 2 {
		// No matches for week 2 (league finished)
		return []*models.Match{}, nil
	}
	return nil, fmt.Errorf("no matches found for league %d week %d", leagueID, week)
}

func (m *mockLeagueDBService) PlayMatch(ctx context.Context, matchID, homeGoals, awayGoals int) error {
	if matchID == 1 {
		return nil // Successful update
	}
	return fmt.Errorf("no scheduled match found with ID %d", matchID)
}

func (m *mockLeagueDBService) UpdateStandings(ctx context.Context, leagueID, homeTeamID, awayTeamID, homeGoals, awayGoals int) error {
	if leagueID == 1 || leagueID == 3 {
		return nil // Successful update
	}
	return fmt.Errorf("failed to update standings")
}

func (m *mockLeagueDBService) AdvanceLeagueWeek(ctx context.Context, leagueID int) error {
	if leagueID == 1 || leagueID == 3 {
		return nil // Successful update
	}
	return fmt.Errorf("no league found with ID %d", leagueID)
}

func (m *mockLeagueDBService) GetStandings(ctx context.Context, leagueID int) ([]models.StandingWithTeam, error) {
	if leagueID == 1 || leagueID == 3 {
		return []models.StandingWithTeam{
			{
				Standing: models.Standing{
					LeagueID:       leagueID,
					TeamID:         1,
					Points:         9,
					Played:         3,
					Wins:           3,
					Draws:          0,
					Losses:         0,
					GoalsFor:       6,
					GoalsAgainst:   2,
					GoalDifference: 4,
				},
				TeamName: "Team A",
			},
			{
				Standing: models.Standing{
					LeagueID:       leagueID,
					TeamID:         2,
					Points:         6,
					Played:         3,
					Wins:           2,
					Draws:          0,
					Losses:         1,
					GoalsFor:       4,
					GoalsAgainst:   3,
					GoalDifference: 1,
				},
				TeamName: "Team B",
			},
		}, nil
	}
	return nil, fmt.Errorf("no standings found for league %d", leagueID)
}

func TestCreateLeagueHandler(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	// Test data
	leagueReq := models.CreateLeagueRequest{
		Name: "Premier League",
	}

	// Convert to JSON
	reqBody, err := json.Marshal(leagueReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/api/leagues/create", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call the handler
	handler.CreateLeagueHandler(w, req)

	// Check status code
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	// Check content type
	expectedContentType := "application/json"
	if contentType := w.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("Expected content type %s, got %s", expectedContentType, contentType)
	}

	// Parse response
	var resp models.LeagueResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response data
	if resp.Name != leagueReq.Name {
		t.Errorf("Expected name %s, got %s", leagueReq.Name, resp.Name)
	}
	if resp.Status != "created" {
		t.Errorf("Expected status 'created', got %s", resp.Status)
	}
	if resp.CurrentWeek != 0 {
		t.Errorf("Expected current_week 0, got %d", resp.CurrentWeek)
	}
	if resp.ID != 1 {
		t.Errorf("Expected ID %d, got %d", 1, resp.ID)
	}
}

func TestCreateLeagueHandler_EmptyName(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	leagueReq := models.CreateLeagueRequest{
		Name: "",
	}

	reqBody, _ := json.Marshal(leagueReq)
	req := httptest.NewRequest(http.MethodPost, "/api/leagues/create", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.CreateLeagueHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateLeagueHandler_InvalidMethod(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodGet, "/api/leagues/create", nil)
	w := httptest.NewRecorder()

	handler.CreateLeagueHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestCreateLeagueHandler_InvalidJSON(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodPost, "/api/leagues/create", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.CreateLeagueHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestInitializeLeagueHandler(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	// Test data
	leagueReq := models.CreateLeagueRequest{
		Name: "Initialized Premier League",
	}

	// Convert to JSON
	reqBody, err := json.Marshal(leagueReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/api/leagues/initialize", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call the handler
	handler.InitializeLeagueHandler(w, req)

	// Check status code
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	// Check content type
	expectedContentType := "application/json"
	if contentType := w.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("Expected content type %s, got %s", expectedContentType, contentType)
	}

	// Parse response
	var resp models.InitializeLeagueResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response data
	if resp.League.Name != leagueReq.Name {
		t.Errorf("Expected league name %s, got %s", leagueReq.Name, resp.League.Name)
	}
	if resp.League.Status != "created" {
		t.Errorf("Expected league status 'created', got %s", resp.League.Status)
	}
	if len(resp.Teams) != 4 {
		t.Errorf("Expected 4 teams, got %d", len(resp.Teams))
	}
	if resp.Teams[0].Name != "Manchester City" {
		t.Errorf("Expected first team 'Manchester City', got %s", resp.Teams[0].Name)
	}
	if resp.Message == "" {
		t.Error("Expected non-empty message")
	}
}

func TestInitializeLeagueHandler_EmptyName(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	leagueReq := models.CreateLeagueRequest{
		Name: "",
	}

	reqBody, _ := json.Marshal(leagueReq)
	req := httptest.NewRequest(http.MethodPost, "/api/leagues/initialize", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.InitializeLeagueHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestInitializeLeagueHandler_InvalidMethod(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodGet, "/api/leagues/initialize", nil)
	w := httptest.NewRecorder()

	handler.InitializeLeagueHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestAddTeamToLeagueHandler(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodPost, "/api/leagues/add-team/1/1", nil)
	w := httptest.NewRecorder()

	handler.AddTeamToLeagueHandler(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	// Parse response
	var resp models.AddTeamToLeagueResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response data
	if resp.League.ID != 1 {
		t.Errorf("Expected league ID 1, got %d", resp.League.ID)
	}
	if resp.Team.ID != 1 {
		t.Errorf("Expected team ID 1, got %d", resp.Team.ID)
	}
	if resp.Message == "" {
		t.Error("Expected non-empty message")
	}
}

func TestAddTeamToLeagueHandler_LeagueNotFound(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodPost, "/api/leagues/add-team/99/1", nil)
	w := httptest.NewRecorder()

	handler.AddTeamToLeagueHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestAddTeamToLeagueHandler_TeamNotFound(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodPost, "/api/leagues/add-team/1/99", nil)
	w := httptest.NewRecorder()

	handler.AddTeamToLeagueHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestAddTeamToLeagueHandler_InvalidLeagueID(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodPost, "/api/leagues/add-team/abc/1", nil)
	w := httptest.NewRecorder()

	handler.AddTeamToLeagueHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAddTeamToLeagueHandler_InvalidTeamID(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodPost, "/api/leagues/add-team/1/abc", nil)
	w := httptest.NewRecorder()

	handler.AddTeamToLeagueHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAddTeamToLeagueHandler_InvalidMethod(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodGet, "/api/leagues/add-team/1/1", nil)
	w := httptest.NewRecorder()

	handler.AddTeamToLeagueHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestAddTeamToLeagueHandler_InvalidPath(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodPost, "/api/leagues/add-team/1", nil)
	w := httptest.NewRecorder()

	handler.AddTeamToLeagueHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestRemoveTeamFromLeagueHandler(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodPost, "/api/leagues/remove-team/1/1", nil)
	w := httptest.NewRecorder()

	handler.RemoveTeamFromLeagueHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Parse response
	var resp models.RemoveTeamFromLeagueResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response data
	if resp.League.ID != 1 {
		t.Errorf("Expected league ID 1, got %d", resp.League.ID)
	}
	if resp.Team.ID != 1 {
		t.Errorf("Expected team ID 1, got %d", resp.Team.ID)
	}
	if resp.Message == "" {
		t.Error("Expected non-empty message")
	}
}

func TestRemoveTeamFromLeagueHandler_TeamNotInLeague(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodPost, "/api/leagues/remove-team/1/2", nil)
	w := httptest.NewRecorder()

	handler.RemoveTeamFromLeagueHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestRemoveTeamFromLeagueHandler_LeagueNotFound(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodPost, "/api/leagues/remove-team/99/1", nil)
	w := httptest.NewRecorder()

	handler.RemoveTeamFromLeagueHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestRemoveTeamFromLeagueHandler_TeamNotFound(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodPost, "/api/leagues/remove-team/1/99", nil)
	w := httptest.NewRecorder()

	handler.RemoveTeamFromLeagueHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestRemoveTeamFromLeagueHandler_InvalidLeagueID(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodPost, "/api/leagues/remove-team/abc/1", nil)
	w := httptest.NewRecorder()

	handler.RemoveTeamFromLeagueHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestRemoveTeamFromLeagueHandler_InvalidTeamID(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodPost, "/api/leagues/remove-team/1/abc", nil)
	w := httptest.NewRecorder()

	handler.RemoveTeamFromLeagueHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestRemoveTeamFromLeagueHandler_InvalidMethod(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodGet, "/api/leagues/remove-team/1/1", nil)
	w := httptest.NewRecorder()

	handler.RemoveTeamFromLeagueHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestRemoveTeamFromLeagueHandler_InvalidPath(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodPost, "/api/leagues/remove-team/1", nil)
	w := httptest.NewRecorder()

	handler.RemoveTeamFromLeagueHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestStartLeagueHandler(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodPost, "/api/leagues/start/1", nil)
	w := httptest.NewRecorder()

	handler.StartLeagueHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Parse response
	var resp models.StartLeagueResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response data
	if resp.League.ID != 1 {
		t.Errorf("Expected league ID 1, got %d", resp.League.ID)
	}
	if resp.League.Status != "started" {
		t.Errorf("Expected league status 'started', got %s", resp.League.Status)
	}
	if resp.TeamsCount != 2 {
		t.Errorf("Expected 2 teams, got %d", resp.TeamsCount)
	}
	// 2 teams = 2 * (2-1) = 2 matches total
	if resp.MatchesCount != 2 {
		t.Errorf("Expected 2 matches, got %d", resp.MatchesCount)
	}
	if resp.Message == "" {
		t.Error("Expected non-empty message")
	}
}

func TestStartLeagueHandler_LeagueNotFound(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodPost, "/api/leagues/start/99", nil)
	w := httptest.NewRecorder()

	handler.StartLeagueHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestStartLeagueHandler_InvalidLeagueID(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodPost, "/api/leagues/start/abc", nil)
	w := httptest.NewRecorder()

	handler.StartLeagueHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestStartLeagueHandler_InvalidMethod(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodGet, "/api/leagues/start/1", nil)
	w := httptest.NewRecorder()

	handler.StartLeagueHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestStartLeagueHandler_InvalidPath(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodPost, "/api/leagues/start", nil)
	w := httptest.NewRecorder()

	handler.StartLeagueHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAdvanceWeekHandler(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodPost, "/api/leagues/advance-week/3", nil)
	w := httptest.NewRecorder()

	handler.AdvanceWeekHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Parse response
	var resp models.AdvanceWeekResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response data
	if resp.League.ID != 3 {
		t.Errorf("Expected league ID 3, got %d", resp.League.ID)
	}
	if resp.WeekAdvanced != 1 {
		t.Errorf("Expected week advanced 1, got %d", resp.WeekAdvanced)
	}
	if len(resp.MatchesPlayed) != 1 {
		t.Errorf("Expected 1 match played, got %d", len(resp.MatchesPlayed))
	}
	if resp.Message == "" {
		t.Error("Expected non-empty message")
	}
}

func TestAdvanceWeekHandler_LeagueNotFound(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodPost, "/api/leagues/advance-week/99", nil)
	w := httptest.NewRecorder()

	handler.AdvanceWeekHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestAdvanceWeekHandler_LeagueNotStarted(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodPost, "/api/leagues/advance-week/2", nil)
	w := httptest.NewRecorder()

	handler.AdvanceWeekHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAdvanceWeekHandler_InvalidLeagueID(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodPost, "/api/leagues/advance-week/abc", nil)
	w := httptest.NewRecorder()

	handler.AdvanceWeekHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAdvanceWeekHandler_InvalidMethod(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodGet, "/api/leagues/advance-week/1", nil)
	w := httptest.NewRecorder()

	handler.AdvanceWeekHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestAdvanceWeekHandler_InvalidPath(t *testing.T) {
	handler := NewLeagueHandler(&mockLeagueDBService{})

	req := httptest.NewRequest(http.MethodPost, "/api/leagues/advance-week", nil)
	w := httptest.NewRecorder()

	handler.AdvanceWeekHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
