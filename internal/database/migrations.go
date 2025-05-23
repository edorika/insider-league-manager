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
