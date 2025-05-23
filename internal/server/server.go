package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"insider-league-manager/internal/database"
	"insider-league-manager/internal/handlers"
)

type Server struct {
	port int

	db          database.Service
	teamHandler *handlers.TeamHandler
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	db := database.New()

	// Initialize database tables
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := db.InitializeTables(ctx); err != nil {
		panic(fmt.Sprintf("failed to initialize database tables: %v", err))
	}

	NewServer := &Server{
		port:        port,
		db:          db,
		teamHandler: handlers.NewTeamHandler(db),
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
