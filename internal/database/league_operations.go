package database

import (
	"context"
	"database/sql"
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

// RemoveTeamFromLeague removes a team from a league and their standings
func (s *service) RemoveTeamFromLeague(ctx context.Context, leagueID, teamID int) error {
	// First, check if the team is actually in the league
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM league_teams WHERE league_id = $1 AND team_id = $2)`
	err := s.db.QueryRowContext(ctx, checkQuery, leagueID, teamID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if team %d exists in league %d: %w", teamID, leagueID, err)
	}

	if !exists {
		return fmt.Errorf("team %d is not in league %d", teamID, leagueID)
	}

	// Remove from standings first (due to foreign key constraints)
	deleteStandingsQuery := `DELETE FROM standings WHERE league_id = $1 AND team_id = $2`
	_, err = s.db.ExecContext(ctx, deleteStandingsQuery, leagueID, teamID)
	if err != nil {
		return fmt.Errorf("failed to remove standings for team %d in league %d: %w", teamID, leagueID, err)
	}

	// Remove from league_teams
	deleteLeagueTeamQuery := `DELETE FROM league_teams WHERE league_id = $1 AND team_id = $2`
	result, err := s.db.ExecContext(ctx, deleteLeagueTeamQuery, leagueID, teamID)
	if err != nil {
		return fmt.Errorf("failed to remove team %d from league %d: %w", teamID, leagueID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after removing team %d from league %d: %w", teamID, leagueID, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no team found with ID %d in league %d", teamID, leagueID)
	}

	return nil
}

// GetTeamsInLeague retrieves all teams that are part of a specific league
func (s *service) GetTeamsInLeague(ctx context.Context, leagueID int) ([]*models.Team, error) {
	query := `
		SELECT t.id, t.name, t.strength 
		FROM teams t
		INNER JOIN league_teams lt ON t.id = lt.team_id
		WHERE lt.league_id = $1
		ORDER BY t.name
	`

	rows, err := s.db.QueryContext(ctx, query, leagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to query teams in league %d: %w", leagueID, err)
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

// CreateMatch creates a new match in the database
func (s *service) CreateMatch(ctx context.Context, match *models.Match) (*models.Match, error) {
	insertQuery := `
		INSERT INTO matches (league_id, home_team_id, away_team_id, week, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, league_id, home_team_id, away_team_id, week, home_goals, away_goals, status, played_at, created_at
	`

	createdMatch := &models.Match{}
	err := s.db.QueryRowContext(
		ctx,
		insertQuery,
		match.LeagueID,
		match.HomeTeamID,
		match.AwayTeamID,
		match.Week,
		match.Status,
	).Scan(
		&createdMatch.ID,
		&createdMatch.LeagueID,
		&createdMatch.HomeTeamID,
		&createdMatch.AwayTeamID,
		&createdMatch.Week,
		&createdMatch.HomeGoals,
		&createdMatch.AwayGoals,
		&createdMatch.Status,
		&createdMatch.PlayedAt,
		&createdMatch.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create match: %w", err)
	}

	return createdMatch, nil
}

// UpdateLeagueStatus updates the status of a league
func (s *service) UpdateLeagueStatus(ctx context.Context, leagueID int, status string) error {
	updateQuery := `UPDATE leagues SET status = $1 WHERE id = $2`

	result, err := s.db.ExecContext(ctx, updateQuery, status, leagueID)
	if err != nil {
		return fmt.Errorf("failed to update league %d status to %s: %w", leagueID, status, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after updating league %d: %w", leagueID, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no league found with ID %d", leagueID)
	}

	return nil
}

// GetMatchesByWeekAndLeague retrieves matches for a specific league and week
func (s *service) GetMatchesByWeekAndLeague(ctx context.Context, leagueID, week int) ([]*models.Match, error) {
	query := `
		SELECT id, league_id, home_team_id, away_team_id, week, home_goals, away_goals, status, played_at, created_at
		FROM matches 
		WHERE league_id = $1 AND week = $2
		ORDER BY id
	`

	rows, err := s.db.QueryContext(ctx, query, leagueID, week)
	if err != nil {
		return nil, fmt.Errorf("failed to query matches for league %d week %d: %w", leagueID, week, err)
	}
	defer rows.Close()

	var matches []*models.Match
	for rows.Next() {
		match := &models.Match{}
		err := rows.Scan(
			&match.ID,
			&match.LeagueID,
			&match.HomeTeamID,
			&match.AwayTeamID,
			&match.Week,
			&match.HomeGoals,
			&match.AwayGoals,
			&match.Status,
			&match.PlayedAt,
			&match.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan match: %w", err)
		}
		matches = append(matches, match)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over matches: %w", err)
	}

	return matches, nil
}

// PlayMatch updates a match with results and marks it as played
func (s *service) PlayMatch(ctx context.Context, matchID, homeGoals, awayGoals int) error {
	updateQuery := `
		UPDATE matches 
		SET home_goals = $1, away_goals = $2, status = 'played', played_at = NOW()
		WHERE id = $3
	`

	result, err := s.db.ExecContext(ctx, updateQuery, homeGoals, awayGoals, matchID)
	if err != nil {
		return fmt.Errorf("failed to update match %d: %w", matchID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after updating match %d: %w", matchID, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no match found with ID %d", matchID)
	}

	return nil
}

// UpdateStandings updates team standings after a match
func (s *service) UpdateStandings(ctx context.Context, leagueID, homeTeamID, awayTeamID, homeGoals, awayGoals int) error {
	// Determine match result
	var homePoints, awayPoints int
	var homeWins, homeDraws, homeLosses int
	var awayWins, awayDraws, awayLosses int

	if homeGoals > awayGoals {
		// Home team wins
		homePoints = 3
		awayPoints = 0
		homeWins = 1
		awayLosses = 1
	} else if homeGoals < awayGoals {
		// Away team wins
		homePoints = 0
		awayPoints = 3
		homeLosses = 1
		awayWins = 1
	} else {
		// Draw
		homePoints = 1
		awayPoints = 1
		homeDraws = 1
		awayDraws = 1
	}

	// Update home team standings
	homeUpdateQuery := `
		UPDATE standings 
		SET points = points + $1,
		    played = played + 1,
		    wins = wins + $2,
		    draws = draws + $3,
		    losses = losses + $4,
		    goals_for = goals_for + $5,
		    goals_against = goals_against + $6,
		    goal_difference = goals_for + $5 - (goals_against + $6)
		WHERE league_id = $7 AND team_id = $8
	`

	_, err := s.db.ExecContext(ctx, homeUpdateQuery,
		homePoints, homeWins, homeDraws, homeLosses, homeGoals, awayGoals, leagueID, homeTeamID)
	if err != nil {
		return fmt.Errorf("failed to update home team %d standings: %w", homeTeamID, err)
	}

	// Update away team standings
	awayUpdateQuery := `
		UPDATE standings 
		SET points = points + $1,
		    played = played + 1,
		    wins = wins + $2,
		    draws = draws + $3,
		    losses = losses + $4,
		    goals_for = goals_for + $5,
		    goals_against = goals_against + $6,
		    goal_difference = goals_for + $5 - (goals_against + $6)
		WHERE league_id = $7 AND team_id = $8
	`

	_, err = s.db.ExecContext(ctx, awayUpdateQuery,
		awayPoints, awayWins, awayDraws, awayLosses, awayGoals, homeGoals, leagueID, awayTeamID)
	if err != nil {
		return fmt.Errorf("failed to update away team %d standings: %w", awayTeamID, err)
	}

	return nil
}

// AdvanceLeagueWeek increments the current week of a league
func (s *service) AdvanceLeagueWeek(ctx context.Context, leagueID int) error {
	updateQuery := `UPDATE leagues SET current_week = current_week + 1 WHERE id = $1`

	result, err := s.db.ExecContext(ctx, updateQuery, leagueID)
	if err != nil {
		return fmt.Errorf("failed to advance week for league %d: %w", leagueID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after advancing league %d week: %w", leagueID, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no league found with ID %d", leagueID)
	}

	return nil
}

// GetStandings retrieves league standings sorted by points and goal difference
func (s *service) GetStandings(ctx context.Context, leagueID int) ([]models.StandingWithTeam, error) {
	query := `
		SELECT s.league_id, s.team_id, s.points, s.played, s.wins, s.draws, s.losses, 
		       s.goals_for, s.goals_against, s.goal_difference, t.name as team_name
		FROM standings s
		INNER JOIN teams t ON s.team_id = t.id
		WHERE s.league_id = $1
		ORDER BY s.points DESC, s.goal_difference DESC, s.goals_for DESC, t.name ASC
	`

	rows, err := s.db.QueryContext(ctx, query, leagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to query standings for league %d: %w", leagueID, err)
	}
	defer rows.Close()

	var standings []models.StandingWithTeam
	for rows.Next() {
		var standing models.StandingWithTeam
		err := rows.Scan(
			&standing.LeagueID,
			&standing.TeamID,
			&standing.Points,
			&standing.Played,
			&standing.Wins,
			&standing.Draws,
			&standing.Losses,
			&standing.GoalsFor,
			&standing.GoalsAgainst,
			&standing.GoalDifference,
			&standing.TeamName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan standing: %w", err)
		}
		standings = append(standings, standing)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over standings: %w", err)
	}

	return standings, nil
}

// GetMatchByID retrieves a match by its ID
func (s *service) GetMatchByID(ctx context.Context, matchID int) (*models.Match, error) {
	query := `
		SELECT id, league_id, home_team_id, away_team_id, week, home_goals, away_goals, status, played_at, created_at
		FROM matches 
		WHERE id = $1
	`

	var match models.Match
	err := s.db.QueryRowContext(ctx, query, matchID).Scan(
		&match.ID,
		&match.LeagueID,
		&match.HomeTeamID,
		&match.AwayTeamID,
		&match.Week,
		&match.HomeGoals,
		&match.AwayGoals,
		&match.Status,
		&match.PlayedAt,
		&match.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get match by ID %d: %w", matchID, err)
	}

	return &match, nil
}

// EditMatch updates match result and recalculates standings
func (s *service) EditMatch(ctx context.Context, matchID, newHomeGoals, newAwayGoals int) error {
	// Start a transaction to ensure all operations succeed or fail together
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get the current match details
	getMatchQuery := `
		SELECT league_id, home_team_id, away_team_id, home_goals, away_goals, status
		FROM matches 
		WHERE id = $1
	`

	var leagueID, homeTeamID, awayTeamID int
	var oldHomeGoals, oldAwayGoals *int
	var status string

	err = tx.QueryRowContext(ctx, getMatchQuery, matchID).Scan(
		&leagueID, &homeTeamID, &awayTeamID, &oldHomeGoals, &oldAwayGoals, &status,
	)
	if err != nil {
		return fmt.Errorf("failed to get match details: %w", err)
	}

	// Check if match can be edited (must be played)
	if status != "played" {
		return fmt.Errorf("can only edit matches with 'played' status, current status: %s", status)
	}

	if oldHomeGoals == nil || oldAwayGoals == nil {
		return fmt.Errorf("match has no existing result to edit")
	}

	// Update the match with new results
	updateMatchQuery := `
		UPDATE matches 
		SET home_goals = $1, away_goals = $2 
		WHERE id = $3
	`

	_, err = tx.ExecContext(ctx, updateMatchQuery, newHomeGoals, newAwayGoals, matchID)
	if err != nil {
		return fmt.Errorf("failed to update match: %w", err)
	}

	// Reverse the old standings effect
	err = s.reverseStandingsEffect(ctx, tx, leagueID, homeTeamID, awayTeamID, *oldHomeGoals, *oldAwayGoals)
	if err != nil {
		return fmt.Errorf("failed to reverse old standings: %w", err)
	}

	// Apply the new standings effect
	err = s.applyStandingsEffect(ctx, tx, leagueID, homeTeamID, awayTeamID, newHomeGoals, newAwayGoals)
	if err != nil {
		return fmt.Errorf("failed to apply new standings: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// reverseStandingsEffect removes the effect of the old match result from standings
func (s *service) reverseStandingsEffect(ctx context.Context, tx *sql.Tx, leagueID, homeTeamID, awayTeamID, homeGoals, awayGoals int) error {
	// Calculate what needs to be reversed
	var homePoints, awayPoints int
	var homeWins, homeDraws, homeLosses int
	var awayWins, awayDraws, awayLosses int

	if homeGoals > awayGoals {
		// Home team won
		homePoints = 3
		awayPoints = 0
		homeWins = 1
		awayLosses = 1
	} else if homeGoals < awayGoals {
		// Away team won
		homePoints = 0
		awayPoints = 3
		homeLosses = 1
		awayWins = 1
	} else {
		// Draw
		homePoints = 1
		awayPoints = 1
		homeDraws = 1
		awayDraws = 1
	}

	// Reverse home team standings
	homeQuery := `
		UPDATE standings 
		SET points = points - $1,
		    played = played - 1,
		    wins = wins - $2,
		    draws = draws - $3,
		    losses = losses - $4,
		    goals_for = goals_for - $5,
		    goals_against = goals_against - $6,
		    goal_difference = goal_difference - ($5 - $6)
		WHERE league_id = $7 AND team_id = $8
	`

	_, err := tx.ExecContext(ctx, homeQuery,
		homePoints, homeWins, homeDraws, homeLosses,
		homeGoals, awayGoals, leagueID, homeTeamID)
	if err != nil {
		return fmt.Errorf("failed to reverse home team standings: %w", err)
	}

	// Reverse away team standings
	awayQuery := `
		UPDATE standings 
		SET points = points - $1,
		    played = played - 1,
		    wins = wins - $2,
		    draws = draws - $3,
		    losses = losses - $4,
		    goals_for = goals_for - $5,
		    goals_against = goals_against - $6,
		    goal_difference = goal_difference - ($5 - $6)
		WHERE league_id = $7 AND team_id = $8
	`

	_, err = tx.ExecContext(ctx, awayQuery,
		awayPoints, awayWins, awayDraws, awayLosses,
		awayGoals, homeGoals, leagueID, awayTeamID)
	if err != nil {
		return fmt.Errorf("failed to reverse away team standings: %w", err)
	}

	return nil
}

// applyStandingsEffect applies the effect of the new match result to standings
func (s *service) applyStandingsEffect(ctx context.Context, tx *sql.Tx, leagueID, homeTeamID, awayTeamID, homeGoals, awayGoals int) error {
	// Calculate what needs to be applied
	var homePoints, awayPoints int
	var homeWins, homeDraws, homeLosses int
	var awayWins, awayDraws, awayLosses int

	if homeGoals > awayGoals {
		// Home team won
		homePoints = 3
		awayPoints = 0
		homeWins = 1
		awayLosses = 1
	} else if homeGoals < awayGoals {
		// Away team won
		homePoints = 0
		awayPoints = 3
		homeLosses = 1
		awayWins = 1
	} else {
		// Draw
		homePoints = 1
		awayPoints = 1
		homeDraws = 1
		awayDraws = 1
	}

	// Apply home team standings
	homeQuery := `
		UPDATE standings 
		SET points = points + $1,
		    played = played + 1,
		    wins = wins + $2,
		    draws = draws + $3,
		    losses = losses + $4,
		    goals_for = goals_for + $5,
		    goals_against = goals_against + $6,
		    goal_difference = goal_difference + ($5 - $6)
		WHERE league_id = $7 AND team_id = $8
	`

	_, err := tx.ExecContext(ctx, homeQuery,
		homePoints, homeWins, homeDraws, homeLosses,
		homeGoals, awayGoals, leagueID, homeTeamID)
	if err != nil {
		return fmt.Errorf("failed to apply home team standings: %w", err)
	}

	// Apply away team standings
	awayQuery := `
		UPDATE standings 
		SET points = points + $1,
		    played = played + 1,
		    wins = wins + $2,
		    draws = draws + $3,
		    losses = losses + $4,
		    goals_for = goals_for + $5,
		    goals_against = goals_against + $6,
		    goal_difference = goal_difference + ($5 - $6)
		WHERE league_id = $7 AND team_id = $8
	`

	_, err = tx.ExecContext(ctx, awayQuery,
		awayPoints, awayWins, awayDraws, awayLosses,
		awayGoals, homeGoals, leagueID, awayTeamID)
	if err != nil {
		return fmt.Errorf("failed to apply away team standings: %w", err)
	}

	return nil
}
