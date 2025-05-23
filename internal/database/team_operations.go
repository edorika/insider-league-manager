package database

import (
	"context"
	"fmt"

	"insider-league-manager/internal/models"
)

// CreateTeam creates a new team in the database
func (s *service) CreateTeam(ctx context.Context, req *models.CreateTeamRequest) (*models.Team, error) {
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

// UpdateTeam updates a team in the database
func (s *service) UpdateTeam(ctx context.Context, teamID int, req *models.CreateTeamRequest) (*models.Team, error) {
	updateQuery := `
		UPDATE teams 
		SET name = $1, strength = $2
		WHERE id = $3
		RETURNING id, name, strength
	`

	team := &models.Team{}
	err := s.db.QueryRowContext(
		ctx,
		updateQuery,
		req.Name,
		req.Strength,
		teamID,
	).Scan(
		&team.ID,
		&team.Name,
		&team.Strength,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update team with ID %d: %w", teamID, err)
	}

	return team, nil
}

// DeleteTeam deletes a team from the database
func (s *service) DeleteTeam(ctx context.Context, teamID int) error {
	deleteQuery := `DELETE FROM teams WHERE id = $1`

	result, err := s.db.ExecContext(ctx, deleteQuery, teamID)
	if err != nil {
		return fmt.Errorf("failed to delete team with ID %d: %w", teamID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after deleting team with ID %d: %w", teamID, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no team found with ID %d", teamID)
	}

	return nil
}
