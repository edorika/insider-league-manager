package handlers

import (
	"bytes"
	"context"
	"encoding/json"
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
