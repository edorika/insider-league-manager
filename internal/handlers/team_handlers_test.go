package handlers

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

func (m *mockDBService) InitializeTables(ctx context.Context) error {
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

func (m *mockDBService) UpdateTeam(ctx context.Context, teamID int, req *models.CreateTeamRequest) (*models.Team, error) {
	if teamID == 1 {
		return &models.Team{
			ID:       1,
			Name:     req.Name,
			Strength: req.Strength,
		}, nil
	}
	// Return error for any other ID to simulate not found
	return nil, fmt.Errorf("no rows in result set")
}

func (m *mockDBService) DeleteTeam(ctx context.Context, teamID int) error {
	if teamID == 1 {
		return nil // Successful deletion
	}
	// Return error for any other ID to simulate not found
	return fmt.Errorf("no team found with ID %d", teamID)
}

func TestCreateTeamHandler(t *testing.T) {
	handler := NewTeamHandler(&mockDBService{})

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
	handler.CreateTeamHandler(w, req)

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
	handler := NewTeamHandler(&mockDBService{})

	req := httptest.NewRequest(http.MethodGet, "/api/teams", nil)
	w := httptest.NewRecorder()

	handler.GetAllTeamsHandler(w, req)

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
	handler := NewTeamHandler(&mockDBService{})

	req := httptest.NewRequest(http.MethodGet, "/api/teams/1", nil)
	w := httptest.NewRecorder()

	handler.GetTeamByIDHandler(w, req)

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

func TestUpdateTeamHandler(t *testing.T) {
	handler := NewTeamHandler(&mockDBService{})

	// Test data
	teamReq := models.CreateTeamRequest{
		Name:     "Updated Team",
		Strength: 95,
	}

	// Convert to JSON
	reqBody, err := json.Marshal(teamReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Create request
	req := httptest.NewRequest(http.MethodPut, "/api/teams/1", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call the handler
	handler.UpdateTeamHandler(w, req)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
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

func TestUpdateTeamHandler_NotFound(t *testing.T) {
	handler := NewTeamHandler(&mockDBService{})

	teamReq := models.CreateTeamRequest{
		Name:     "Updated Team",
		Strength: 95,
	}

	reqBody, _ := json.Marshal(teamReq)
	req := httptest.NewRequest(http.MethodPut, "/api/teams/99", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.UpdateTeamHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetTeamByIDHandler_NotFound(t *testing.T) {
	handler := NewTeamHandler(&mockDBService{})

	req := httptest.NewRequest(http.MethodGet, "/api/teams/99", nil)
	w := httptest.NewRecorder()

	handler.GetTeamByIDHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestCreateTeamHandler_EmptyName(t *testing.T) {
	handler := NewTeamHandler(&mockDBService{})

	teamReq := models.CreateTeamRequest{
		Name:     "",
		Strength: 75,
	}

	reqBody, _ := json.Marshal(teamReq)
	req := httptest.NewRequest(http.MethodPost, "/api/teams", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.CreateTeamHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateTeamHandler_EmptyName(t *testing.T) {
	handler := NewTeamHandler(&mockDBService{})

	teamReq := models.CreateTeamRequest{
		Name:     "",
		Strength: 75,
	}

	reqBody, _ := json.Marshal(teamReq)
	req := httptest.NewRequest(http.MethodPut, "/api/teams/1", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.UpdateTeamHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDeleteTeamHandler(t *testing.T) {
	handler := NewTeamHandler(&mockDBService{})

	req := httptest.NewRequest(http.MethodDelete, "/api/teams/1", nil)
	w := httptest.NewRecorder()

	handler.DeleteTeamHandler(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	// Verify response body is empty for 204 No Content
	if w.Body.Len() != 0 {
		t.Errorf("Expected empty response body, got %s", w.Body.String())
	}
}

func TestDeleteTeamHandler_NotFound(t *testing.T) {
	handler := NewTeamHandler(&mockDBService{})

	req := httptest.NewRequest(http.MethodDelete, "/api/teams/99", nil)
	w := httptest.NewRecorder()

	handler.DeleteTeamHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestDeleteTeamHandler_InvalidID(t *testing.T) {
	handler := NewTeamHandler(&mockDBService{})

	req := httptest.NewRequest(http.MethodDelete, "/api/teams/abc", nil)
	w := httptest.NewRecorder()

	handler.DeleteTeamHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDeleteTeamHandler_InvalidMethod(t *testing.T) {
	handler := NewTeamHandler(&mockDBService{})

	req := httptest.NewRequest(http.MethodGet, "/api/teams/1", nil)
	w := httptest.NewRecorder()

	handler.DeleteTeamHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}
