package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"insider-league-manager/internal/models"
)

// Mock database service for testing
type mockDBService struct{}

func (m *mockDBService) Health() map[string]string {
	return map[string]string{"status": "up"}
}

func (m *mockDBService) Close() error {
	return nil
}

func (m *mockDBService) CreateTeam(ctx context.Context, req *models.CreateTeamRequest) (*models.Team, error) {
	return &models.Team{
		ID:       1,
		Name:     req.Name,
		Strength: req.Strength,
	}, nil
}

func (m *mockDBService) GetAllTeams(ctx context.Context) ([]*models.Team, error) {
	return []*models.Team{
		{ID: 1, Name: "Team A", Strength: 85},
		{ID: 2, Name: "Team B", Strength: 90},
	}, nil
}

func (m *mockDBService) GetTeamByID(ctx context.Context, teamID int) (*models.Team, error) {
	if teamID == 1 {
		return &models.Team{
			ID:       1,
			Name:     "Team A",
			Strength: 85,
		}, nil
	}
	// Return error for any other ID to simulate not found
	return nil, fmt.Errorf("no rows in result set")
}

func TestCreateTeamHandler(t *testing.T) {
	// Create a server with mock database
	server := &Server{
		db: &mockDBService{},
	}

	// Test data
	teamReq := models.CreateTeamRequest{
		Name:     "Test Team",
		Strength: 85,
	}

	// Convert to JSON
	reqBody, err := json.Marshal(teamReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/api/teams", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call the handler
	server.createTeamHandler(w, req)

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
	var resp models.TeamResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response data
	if resp.Name != teamReq.Name {
		t.Errorf("Expected name %s, got %s", teamReq.Name, resp.Name)
	}
	if resp.Strength != teamReq.Strength {
		t.Errorf("Expected strength %d, got %d", teamReq.Strength, resp.Strength)
	}
	if resp.ID != 1 {
		t.Errorf("Expected ID %d, got %d", 1, resp.ID)
	}
}

func TestGetAllTeamsHandler(t *testing.T) {
	server := &Server{
		db: &mockDBService{},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/teams", nil)
	w := httptest.NewRecorder()

	server.getAllTeamsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Parse response
	var resp []models.TeamResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(resp) != 2 {
		t.Errorf("Expected 2 teams, got %d", len(resp))
	}

	if resp[0].Name != "Team A" {
		t.Errorf("Expected first team name 'Team A', got %s", resp[0].Name)
	}
}

func TestGetTeamByIDHandler(t *testing.T) {
	server := &Server{
		db: &mockDBService{},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/teams/1", nil)
	w := httptest.NewRecorder()

	server.getTeamByIDHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Parse response
	var resp models.TeamResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.ID != 1 {
		t.Errorf("Expected ID 1, got %d", resp.ID)
	}
	if resp.Name != "Team A" {
		t.Errorf("Expected name 'Team A', got %s", resp.Name)
	}
}

func TestGetTeamByIDHandler_NotFound(t *testing.T) {
	server := &Server{
		db: &mockDBService{},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/teams/99", nil)
	w := httptest.NewRecorder()

	server.getTeamByIDHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestCreateTeamHandler_InvalidMethod(t *testing.T) {
	server := &Server{
		db: &mockDBService{},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/teams", nil)
	w := httptest.NewRecorder()

	server.createTeamHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestCreateTeamHandler_EmptyName(t *testing.T) {
	server := &Server{
		db: &mockDBService{},
	}

	teamReq := models.CreateTeamRequest{
		Name:     "",
		Strength: 75,
	}

	reqBody, _ := json.Marshal(teamReq)
	req := httptest.NewRequest(http.MethodPost, "/api/teams", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.createTeamHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
