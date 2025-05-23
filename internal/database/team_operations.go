package database

import (
	"context"
	"fmt"

	"insider-league-manager/internal/models"
)

// CreateTeam creates a new team in the database
func (s *service) CreateTeam(ctx context.Context, req *models.CreateTeamRequest) (*models.Team, error) {
	// Ensure teams table exists
	if err := s.ensureTeamsTableExists(ctx); err != nil {
		return nil, err
	}

	// Insert the new team
	insertQuery := `
		INSERT INTO teams (name, strength)
		VALUES ($1, $2)
		RETURNING id, name, strength
	`

	team := &models.Team{}
	err := s.db.QueryRowContext(
		ctx,
		insertQuery,
		req.Name,
		req.Strength,
	).Scan(
		&team.ID,
		&team.Name,
		&team.Strength,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	return team, nil
}

// GetAllTeams retrieves all teams from the database
func (s *service) GetAllTeams(ctx context.Context) ([]*models.Team, error) {
	// Ensure teams table exists
	if err := s.ensureTeamsTableExists(ctx); err != nil {
		return nil, err
	}

	query := `SELECT id, name, strength FROM teams ORDER BY id`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query teams: %w", err)
	}
	defer rows.Close()

	var teams []*models.Team
	for rows.Next() {
		team := &models.Team{}
		err := rows.Scan(&team.ID, &team.Name, &team.Strength)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team: %w", err)
		}
		teams = append(teams, team)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over teams: %w", err)
	}

	return teams, nil
}

// GetTeamByID retrieves a team by its ID
func (s *service) GetTeamByID(ctx context.Context, teamID int) (*models.Team, error) {
	// Ensure teams table exists
	if err := s.ensureTeamsTableExists(ctx); err != nil {
		return nil, err
	}

	query := `SELECT id, name, strength FROM teams WHERE id = $1`

	team := &models.Team{}
	err := s.db.QueryRowContext(ctx, query, teamID).Scan(
		&team.ID,
		&team.Name,
		&team.Strength,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get team by ID %d: %w", teamID, err)
	}

	return team, nil
}

// ensureTeamsTableExists creates the teams table if it doesn't exist
func (s *service) ensureTeamsTableExists(ctx context.Context) error {
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS teams (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			strength INTEGER NOT NULL DEFAULT 0
		);
	`

	if _, err := s.db.ExecContext(ctx, createTableQuery); err != nil {
		return fmt.Errorf("failed to ensure teams table exists: %w", err)
	}

	return nil
}
