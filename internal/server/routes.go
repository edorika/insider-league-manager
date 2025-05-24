package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/", s.HelloWorldHandler)

	mux.HandleFunc("/health", s.healthHandler)

	// Team routes
	mux.HandleFunc("/api/teams", s.teamsHandler)
	mux.HandleFunc("/api/teams/", s.teamsHandler) // Handle /api/teams/* patterns

	// League routes
	mux.HandleFunc("/api/leagues/create", s.leaguesCreateHandler)
	mux.HandleFunc("/api/leagues/initialize", s.leaguesInitializeHandler)
	mux.HandleFunc("/api/leagues/add-team/", s.leaguesAddTeamHandler)
	mux.HandleFunc("/api/leagues/remove-team/", s.leaguesRemoveTeamHandler)
	mux.HandleFunc("/api/leagues/start/", s.leaguesStartHandler)
	mux.HandleFunc("/api/leagues/advance-week/", s.leaguesAdvanceWeekHandler)
	mux.HandleFunc("/api/leagues/view-matches/", s.leaguesViewMatchesHandler)
	mux.HandleFunc("/api/leagues/play-all-matches/", s.leaguesPlayAllMatchesHandler)
	mux.HandleFunc("/api/leagues/predict-champion/", s.leaguesPredictChampionHandler)
	mux.HandleFunc("/api/leagues/edit-match/", s.leaguesEditMatchHandler)

	// Wrap the mux with CORS middleware
	return s.corsMiddleware(mux)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Replace "*" with specific origins if needed
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "false") // Set to "true" if credentials are required

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Proceed with the next handler
		next.ServeHTTP(w, r)
	})
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"message": "Hello World"}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(jsonResp); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := json.Marshal(s.db.Health())
	if err != nil {
		http.Error(w, "Failed to marshal health check response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(resp); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

// teamsHandler routes team requests based on method and path
func (s *Server) teamsHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	pathParts := strings.Split(path, "/")

	// Handle /api/teams (exact match)
	if path == "api/teams" {
		switch r.Method {
		case http.MethodPost:
			s.teamHandler.CreateTeamHandler(w, r)
		case http.MethodGet:
			s.teamHandler.GetAllTeamsHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	// Handle /api/teams/{id}
	if len(pathParts) == 3 && pathParts[0] == "api" && pathParts[1] == "teams" {
		switch r.Method {
		case http.MethodGet:
			s.teamHandler.GetTeamByIDHandler(w, r)
		case http.MethodPut:
			s.teamHandler.UpdateTeamHandler(w, r)
		case http.MethodDelete:
			s.teamHandler.DeleteTeamHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	// If we get here, the path doesn't match any known pattern
	http.Error(w, "Not found", http.StatusNotFound)
}

// leaguesCreateHandler handles POST /api/leagues/create
func (s *Server) leaguesCreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.leagueHandler.CreateLeagueHandler(w, r)
}

// leaguesInitializeHandler handles POST /api/leagues/initialize
func (s *Server) leaguesInitializeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.leagueHandler.InitializeLeagueHandler(w, r)
}

// leaguesAddTeamHandler handles POST /api/leagues/add-team/:leagueID/:teamID
func (s *Server) leaguesAddTeamHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.leagueHandler.AddTeamToLeagueHandler(w, r)
}

// leaguesRemoveTeamHandler handles POST /api/leagues/remove-team/:leagueID/:teamID
func (s *Server) leaguesRemoveTeamHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.leagueHandler.RemoveTeamFromLeagueHandler(w, r)
}

// leaguesStartHandler handles POST /api/leagues/start/:leagueID
func (s *Server) leaguesStartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.leagueHandler.StartLeagueHandler(w, r)
}

// leaguesAdvanceWeekHandler handles POST /api/leagues/advance-week/:leagueID
func (s *Server) leaguesAdvanceWeekHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.leagueHandler.AdvanceWeekHandler(w, r)
}

// leaguesViewMatchesHandler handles GET /api/leagues/view-matches/:leagueID
func (s *Server) leaguesViewMatchesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.leagueHandler.ViewMatchesHandler(w, r)
}

// leaguesPlayAllMatchesHandler handles POST /api/leagues/play-all-matches/:leagueID
func (s *Server) leaguesPlayAllMatchesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.leagueHandler.PlayAllMatchesHandler(w, r)
}

// leaguesPredictChampionHandler handles GET /api/leagues/predict-champion/:leagueID
func (s *Server) leaguesPredictChampionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.leagueHandler.PredictChampionHandler(w, r)
}

// leaguesEditMatchHandler handles POST /api/leagues/edit-match/:matchID
func (s *Server) leaguesEditMatchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.leagueHandler.EditMatchHandler(w, r)
}
