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

// LeagueTeam represents the junction table for teams in leagues
type LeagueTeam struct {
	LeagueID int       `json:"league_id"`
	TeamID   int       `json:"team_id"`
	JoinedAt time.Time `json:"joined_at"`
}

// Match represents a match between two teams in a league
type Match struct {
	ID         int        `json:"id"`
	LeagueID   int        `json:"league_id"`
	HomeTeamID int        `json:"home_team_id"`
	AwayTeamID int        `json:"away_team_id"`
	Week       int        `json:"week"`
	HomeGoals  *int       `json:"home_goals"` // nullable until match is played
	AwayGoals  *int       `json:"away_goals"` // nullable until match is played
	Status     string     `json:"status"`     // "scheduled", "played", "cancelled"
	PlayedAt   *time.Time `json:"played_at"`  // nullable until match is played
	CreatedAt  time.Time  `json:"created_at"`
}

// Standing represents team standings in a league
type Standing struct {
	LeagueID       int `json:"league_id"`
	TeamID         int `json:"team_id"`
	Points         int `json:"points"`
	Played         int `json:"played"`
	Wins           int `json:"wins"`
	Draws          int `json:"draws"`
	Losses         int `json:"losses"`
	GoalsFor       int `json:"goals_for"`
	GoalsAgainst   int `json:"goals_against"`
	GoalDifference int `json:"goal_difference"`
}

// StandingWithTeam represents standing with team information
type StandingWithTeam struct {
	Standing
	TeamName string `json:"team_name"`
}

// InitializeLeagueResponse represents the response for league initialization
type InitializeLeagueResponse struct {
	League  LeagueResponse `json:"league"`
	Teams   []Team         `json:"teams"`
	Message string         `json:"message"`
}

// AddTeamToLeagueResponse represents the response for adding a team to a league
type AddTeamToLeagueResponse struct {
	League  LeagueResponse `json:"league"`
	Team    Team           `json:"team"`
	Message string         `json:"message"`
}

// RemoveTeamFromLeagueResponse represents the response for removing a team from a league
type RemoveTeamFromLeagueResponse struct {
	League  LeagueResponse `json:"league"`
	Team    Team           `json:"team"`
	Message string         `json:"message"`
}

// StartLeagueResponse represents the response for starting a league
type StartLeagueResponse struct {
	League       LeagueResponse `json:"league"`
	TeamsCount   int            `json:"teams_count"`
	MatchesCount int            `json:"matches_count"`
	TotalWeeks   int            `json:"total_weeks"`
	Message      string         `json:"message"`
}

// MatchResult represents a played match result
type MatchResult struct {
	Match    Match  `json:"match"`
	HomeTeam string `json:"home_team"`
	AwayTeam string `json:"away_team"`
	Result   string `json:"result"` // e.g. "3-1", "2-2"
}

// AdvanceWeekResponse represents the response for advancing a league week
type AdvanceWeekResponse struct {
	League        LeagueResponse `json:"league"`
	WeekAdvanced  int            `json:"week_advanced"` // The week that was just played
	MatchesPlayed []MatchResult  `json:"matches_played"`
	Message       string         `json:"message"`
}

// ViewMatchesResponse represents the response for viewing matches for the current week
type ViewMatchesResponse struct {
	League      LeagueResponse `json:"league"`
	CurrentWeek int            `json:"current_week"`
	Matches     []MatchResult  `json:"matches"`
	Message     string         `json:"message"`
}

// WeekResult represents match results for a specific week
type WeekResult struct {
	Week    int           `json:"week"`
	Matches []MatchResult `json:"matches"`
}

// PlayAllMatchesResponse represents the response for playing all remaining matches in a league
type PlayAllMatchesResponse struct {
	League             LeagueResponse `json:"league"`
	StartingWeek       int            `json:"starting_week"`
	FinalWeek          int            `json:"final_week"`
	WeeksPlayed        int            `json:"weeks_played"`
	TotalMatchesPlayed int            `json:"total_matches_played"`
	WeekResults        []WeekResult   `json:"week_results"`
	Message            string         `json:"message"`
}
