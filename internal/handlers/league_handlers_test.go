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
	if leagueID == 1 {
		return &models.League{
			ID:          1,
			Name:        "Test League",
			Status:      "created",
			CurrentWeek: 0,
			CreatedAt:   time.Now(),
		}, nil
	}
	// Return error for any other ID to simulate not found
	return nil, fmt.Errorf("no rows in result set")
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
