package database

import (
	"context"
	"fmt"
	"log"
)

// InitializeTables creates all required database tables
func (s *service) InitializeTables(ctx context.Context) error {
	log.Println("Initializing database tables...")

	if err := s.createTeamsTable(ctx); err != nil {
		return fmt.Errorf("failed to create teams table: %w", err)
	}

	if err := s.createLeaguesTable(ctx); err != nil {
		return fmt.Errorf("failed to create leagues table: %w", err)
	}

	if err := s.createLeagueTeamsTable(ctx); err != nil {
		return fmt.Errorf("failed to create league_teams table: %w", err)
	}

	if err := s.createMatchesTable(ctx); err != nil {
		return fmt.Errorf("failed to create matches table: %w", err)
	}

	if err := s.createStandingsTable(ctx); err != nil {
		return fmt.Errorf("failed to create standings table: %w", err)
	}

	log.Println("Database tables initialized successfully")
	return nil
}

// createTeamsTable creates the teams table
func (s *service) createTeamsTable(ctx context.Context) error {
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS teams (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			strength INTEGER NOT NULL DEFAULT 0
		);
	`

	if _, err := s.db.ExecContext(ctx, createTableQuery); err != nil {
		return fmt.Errorf("failed to create teams table: %w", err)
	}

	return nil
}

// createLeaguesTable creates the leagues table
func (s *service) createLeaguesTable(ctx context.Context) error {
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS leagues (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'created',
			current_week INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`

	if _, err := s.db.ExecContext(ctx, createTableQuery); err != nil {
		return fmt.Errorf("failed to create leagues table: %w", err)
	}

	return nil
}

// createLeagueTeamsTable creates the league_teams table
func (s *service) createLeagueTeamsTable(ctx context.Context) error {
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS league_teams (
			league_id INTEGER NOT NULL,
			team_id INTEGER NOT NULL,
			joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (league_id, team_id),
			FOREIGN KEY (league_id) REFERENCES leagues(id) ON DELETE CASCADE,
			FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE
		);
	`

	if _, err := s.db.ExecContext(ctx, createTableQuery); err != nil {
		return fmt.Errorf("failed to create league_teams table: %w", err)
	}

	return nil
}

// createMatchesTable creates the matches table
func (s *service) createMatchesTable(ctx context.Context) error {
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS matches (
			id SERIAL PRIMARY KEY,
			league_id INTEGER NOT NULL,
			home_team_id INTEGER NOT NULL,
			away_team_id INTEGER NOT NULL,
			week INTEGER NOT NULL,
			home_goals INTEGER,
			away_goals INTEGER,
			status VARCHAR(20) NOT NULL DEFAULT 'scheduled',
			played_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (league_id) REFERENCES leagues(id) ON DELETE CASCADE,
			FOREIGN KEY (home_team_id) REFERENCES teams(id) ON DELETE CASCADE,
			FOREIGN KEY (away_team_id) REFERENCES teams(id) ON DELETE CASCADE,
			CHECK (home_team_id != away_team_id)
		);
	`

	if _, err := s.db.ExecContext(ctx, createTableQuery); err != nil {
		return fmt.Errorf("failed to create matches table: %w", err)
	}

	return nil
}

// createStandingsTable creates the standings table
func (s *service) createStandingsTable(ctx context.Context) error {
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS standings (
			league_id INTEGER NOT NULL,
			team_id INTEGER NOT NULL,
			points INTEGER NOT NULL DEFAULT 0,
			played INTEGER NOT NULL DEFAULT 0,
			wins INTEGER NOT NULL DEFAULT 0,
			draws INTEGER NOT NULL DEFAULT 0,
			losses INTEGER NOT NULL DEFAULT 0,
			goals_for INTEGER NOT NULL DEFAULT 0,
			goals_against INTEGER NOT NULL DEFAULT 0,
			goal_difference INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (league_id, team_id),
			FOREIGN KEY (league_id) REFERENCES leagues(id) ON DELETE CASCADE,
			FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE
		);
	`

	if _, err := s.db.ExecContext(ctx, createTableQuery); err != nil {
		return fmt.Errorf("failed to create standings table: %w", err)
	}

	return nil
}
