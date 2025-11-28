package main

import (
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go-finance-advisor/internal/api/handlers"
	"go-finance-advisor/internal/infrastructure/storage/postgres"
)

var (
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
)

func main() {
	log.Printf("Starting Finance Advisor API (Version: %s, Build Time: %s, Commit: %s)", version, buildTime, gitCommit)

	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Use local Postgres from docker-compose
		dbURL = "postgres://user:password@localhost:5432/finance?sslmode=disable"
	}

	db, err := postgres.New(dbURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize handlers
	h := handlers.New(db)

	// Start background ticker loop
	go h.RunTickerLoop()

	// CORS middleware wrapper
	corsMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next(w, r)
		}
	}

	// Router setup
	http.HandleFunc("/api/health", corsMiddleware(h.Health))
	http.HandleFunc("/api/signup", corsMiddleware(h.Signup))
	http.HandleFunc("/api/login", corsMiddleware(h.Login))
	http.HandleFunc("/api/me", corsMiddleware(h.Me))
	http.HandleFunc("/api/auth/demo", corsMiddleware(h.CreateDemoAccount))

	http.HandleFunc("/api/holdings", corsMiddleware(h.HandleHoldings))
	http.HandleFunc("/api/liabilities", corsMiddleware(h.HandleLiabilities))
	http.HandleFunc("/api/documents", corsMiddleware(h.HandleDocuments))
	http.HandleFunc("/api/profile", corsMiddleware(h.HandleProfile))

	http.HandleFunc("/api/ticker", corsMiddleware(h.HandleTicker))
	http.HandleFunc("/api/overview", corsMiddleware(h.HandleOverview))
	http.HandleFunc("/api/performance", corsMiddleware(h.HandlePerformance))
	http.HandleFunc("/api/allocation", corsMiddleware(h.HandleAllocation))
	http.HandleFunc("/api/transactions", corsMiddleware(h.HandleTransactions))
	http.HandleFunc("/api/budgets", corsMiddleware(h.HandleBudgets))
	http.HandleFunc("/api/goals", corsMiddleware(h.HandleGoals))
	http.HandleFunc("/api/export", corsMiddleware(h.HandleExport))

	http.HandleFunc("/ws/ticker", h.HandleWebSocketTicker)

	http.Handle("/metrics", promhttp.Handler())

	// Static files - serve from frontend directory
	// In Docker, we might map /app/frontend. Locally, it's ../frontend (since we are in backend/cmd/api)
	// We'll try to find the frontend directory.
	frontendDir := "../frontend"
	if _, err := os.Stat(frontendDir); os.IsNotExist(err) {
		// Try going up two levels if running from backend/cmd/api
		if _, err := os.Stat("../../frontend"); err == nil {
			frontendDir = "../../frontend"
		} else if _, err := os.Stat("../../../frontend"); err == nil {
			frontendDir = "../../../frontend"
		}
	}

	fs := http.FileServer(http.Dir(frontendDir))
	http.Handle("/", fs)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Server listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
