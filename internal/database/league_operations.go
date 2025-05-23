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

// GetDefaultTeams retrieves the 4 default teams for league initialization
func (s *service) GetDefaultTeams(ctx context.Context) ([]*models.Team, error) {
	query := `
		SELECT id, name, strength 
		FROM teams 
		WHERE name IN ('Manchester City', 'Liverpool FC', 'Chelsea FC', 'Arsenal FC')
		ORDER BY name
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query default teams: %w", err)
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

	if len(teams) != 4 {
		return nil, fmt.Errorf("expected 4 default teams, found %d", len(teams))
	}

	return teams, nil
}

// AddTeamToLeague adds a team to a league
func (s *service) AddTeamToLeague(ctx context.Context, leagueID, teamID int) error {
	insertQuery := `
		INSERT INTO league_teams (league_id, team_id)
		VALUES ($1, $2)
		ON CONFLICT (league_id, team_id) DO NOTHING
	`

	_, err := s.db.ExecContext(ctx, insertQuery, leagueID, teamID)
	if err != nil {
		return fmt.Errorf("failed to add team %d to league %d: %w", teamID, leagueID, err)
	}

	return nil
}

// InitializeStanding creates initial standing entry for a team in a league
func (s *service) InitializeStanding(ctx context.Context, leagueID, teamID int) error {
	insertQuery := `
		INSERT INTO standings (league_id, team_id, points, played, wins, draws, losses, goals_for, goals_against, goal_difference)
		VALUES ($1, $2, 0, 0, 0, 0, 0, 0, 0, 0)
		ON CONFLICT (league_id, team_id) DO NOTHING
	`

	_, err := s.db.ExecContext(ctx, insertQuery, leagueID, teamID)
	if err != nil {
		return fmt.Errorf("failed to initialize standing for team %d in league %d: %w", teamID, leagueID, err)
	}

	return nil
}

// GetLeagueByID retrieves a league by its ID
func (s *service) GetLeagueByID(ctx context.Context, leagueID int) (*models.League, error) {
	query := `SELECT id, name, status, current_week, created_at FROM leagues WHERE id = $1`

	league := &models.League{}
	err := s.db.QueryRowContext(ctx, query, leagueID).Scan(
		&league.ID,
		&league.Name,
		&league.Status,
		&league.CurrentWeek,
		&league.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get league by ID %d: %w", leagueID, err)
	}

	return league, nil
}
