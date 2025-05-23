package models

// Team represents a sports team in the league
type Team struct {
	ID       int    `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`
	Strength int    `json:"strength" db:"strength"`
}

// CreateTeamRequest represents the request payload for creating a team
type CreateTeamRequest struct {
	Name     string `json:"name" validate:"required"`
	Strength int    `json:"strength"`
}

// TeamResponse represents the response payload for team operations
type TeamResponse struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Strength int    `json:"strength"`
}
