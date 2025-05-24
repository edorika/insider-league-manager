# Insider League Manager

A comprehensive football league management system built with Go and PostgreSQL. This application allows you to create and manage football leagues, schedule matches, simulate games, and predict championship outcomes using Monte Carlo simulations.

## 🌐 Live Deployment

The application is deployed and accessible at: **http://31.97.35.211:8080/**

## 🚀 Features

- **Team Management**: Manage football teams with strength ratings
- **League Creation**: Create and manage football leagues
- **Match Scheduling**: Automatic fixture generation for leagues
- **Match Simulation**: Realistic match simulation based on team strengths
- **Match Editing**: Modify match results and see updated standings
- **Championship Prediction**: Monte Carlo simulation for championship probabilities
- **RESTful API**: Complete REST API for all operations

## 📋 Prerequisites

- Go 1.24.3 or higher
- Docker and Docker Compose
- Make (for using Makefile commands)

## 🛠️ Local Development Setup

### Docker
```bash
# Start complete application (database + API)
make docker-up

# Start only database
make db-up

# View logs
make docker-logs

# Stop services
make docker-down

# Clean up (remove volumes)
make docker-clean
```

## 🐳 Docker Commands

| Command | Description |
|---------|-------------|
| `make docker-up` | Start complete application |
| `make docker-stop` | Stop services (keep containers) |
| `make docker-down` | Stop and remove containers |
| `make docker-restart` | Restart all services |
| `make docker-clean` | Remove everything including volumes |
| `make docker-rebuild` | Rebuild and restart API service |
| `make docker-status` | Show service status |
| `make docker-logs` | View all logs |
| `make docker-logs-api` | View API logs only |
| `make docker-logs-db` | View database logs only |
| `make db-up` | Start database only |
| `make db-down` | Stop database only |

## 📡 API Endpoints

### Teams
- `GET /api/teams` - Get all teams
- `GET /api/teams/{id}` - Get team by ID

### Leagues
- `POST /api/leagues` - Create new league
- `GET /api/leagues` - Get all leagues
- `GET /api/leagues/{id}` - Get league by ID
- `DELETE /api/leagues/{id}` - Delete league

### Matches
- `POST /api/leagues/advance-week/{id}` - Advance league to next week
- `GET /api/leagues/view-matches/{id}` - View matches for current week
- `POST /api/leagues/edit-match/{match_id}` - Edit match result
- `GET /api/leagues/predict-champion/{id}` - Get championship predictions

### Example Usage
```bash
# Create a new league
curl -X POST "http://localhost:8080/api/leagues" \
  -H "Content-Type: application/json" \
  -d '{"name": "Premier League 2024"}'

# Advance to next week and simulate matches
curl -X POST "http://localhost:8080/api/leagues/advance-week/1"

# Get championship predictions
curl -X GET "http://localhost:8080/api/leagues/predict-champion/1"

# Edit a match result
curl -X POST "http://localhost:8080/api/leagues/edit-match/1" \
  -H "Content-Type: application/json" \
  -d '{"home_goals": 3, "away_goals": 1}'
```

## 🗃️ Database

The application uses PostgreSQL with the following main tables:
- `teams` - Team information and strength ratings
- `leagues` - League configurations
- `matches` - Match fixtures and results
- `league_standings` - Real-time league standings

Default teams included:
- Manchester City (Strength: 88)
- Liverpool FC (Strength: 86) 
- Chelsea FC (Strength: 84)
- Arsenal FC (Strength: 82)

## ⚙️ Environment Variables

Create a `.env` file in the project root:
```env
BLUEPRINT_DB_HOST=localhost
BLUEPRINT_DB_PORT=5432
BLUEPRINT_DB_DATABASE=insider_league_manager
BLUEPRINT_DB_USERNAME=postgres
BLUEPRINT_DB_PASSWORD=password123
BLUEPRINT_DB_SCHEMA=public
PORT=8080
```

## 🧪 Testing

```bash
# Run all tests
make test

# Run integration tests
make itest

# Test with live API
curl http://localhost:8080/api/teams
```

## 🎯 Match Simulation Algorithm

The application uses a sophisticated match simulation system:
- **Team Strength**: Each team has a strength rating (0-100)
- **Goal Expectancy**: Calculated based on relative team strengths
- **Random Generation**: Poisson distribution for realistic goal scoring
- **Home Advantage**: Subtle home field advantage in calculations

## 🏆 Championship Prediction

Monte Carlo simulation runs 10,000 scenarios to predict championship probabilities:
- Simulates remaining matches based on team strengths
- Calculates final standings for each simulation
- Provides percentage probability for each team winning

**Live Demo**: http://31.97.35.211:8080/
**Local Development**: http://localhost:8080/
