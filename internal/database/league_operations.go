package database

import (
	"context"
	"fmt"

	"insider-league-manager/internal/models"
)

// CreateLeague creates a new league in the database
func (s *service) CreateLeague(ctx context.Context, req *models.CreateLeagueRequest) (*models.League, error) {
	// Insert the new league
	insertQuery := `
		INSERT INTO leagues (name, status, current_week)
		VALUES ($1, $2, $3)
		RETURNING id, name, status, current_week, created_at
	`

	league := &models.League{}
	err := s.db.QueryRowContext(
		ctx,
		insertQuery,
		req.Name,
		"created", // Default status
		0,         // Default current_week
	).Scan(
		&league.ID,
		&league.Name,
		&league.Status,
		&league.CurrentWeek,
		&league.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create league: %w", err)
	}

	return league, nil
}
