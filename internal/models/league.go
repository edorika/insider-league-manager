package models

import "time"

// League represents a league in the database
type League struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`       // "created", "started", "finished"
	CurrentWeek int       `json:"current_week"` // Current week of the league
	CreatedAt   time.Time `json:"created_at"`
}

// CreateLeagueRequest represents the request payload for creating a league
type CreateLeagueRequest struct {
	Name string `json:"name"`
}

// LeagueResponse represents the response format for league operations
type LeagueResponse struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	CurrentWeek int       `json:"current_week"`
	CreatedAt   time.Time `json:"created_at"`
}
