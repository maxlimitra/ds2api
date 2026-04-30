package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Version is set at build time via ldflags
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

func main() {
	// Load .env file if present (ignored in production where env vars are set directly)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Default port changed to 8080 to avoid conflicts with other local services on 3000
	port := getEnv("PORT", "8080")
	host := getEnv("HOST", "127.0.0.1") // personal preference: bind to localhost only by default

	// Validate port is a valid number
	if _, err := strconv.Atoi(port); err != nil {
		log.Fatalf("Invalid PORT value: %s", port)
	}

	addr := fmt.Sprintf("%s:%s", host, port)

	log.Printf("ds2api %s (commit: %s, built: %s)", Version, Commit, BuildDate)
	log.Printf("Starting server on %s", addr)

	router := setupRouter()

	// Added read/write timeouts to avoid hanging connections
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * 1e9, // 10 seconds in nanoseconds (time.Duration)
		WriteTimeout: 10 * 1e9,
		IdleTimeout:  60 * 1e9,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// setupRouter initialises and returns the HTTP router with all routes registered.
func setupRouter() http.Handler {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", healthHandler)

	// Version endpoint
	mux.HandleFunc("/version", versionHandler)

	return mux
}

// healthHandler responds with a simple JSON health status.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok"}`)
}

// versionHandler responds with build version information.
func versionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"version":%q,"commit":%q,"buildDate":%q}`, Version, Commit, BuildDate)
}

// getEnv returns the value of the environment variable named by key,
// or fallback if the variable is not set or empty.
func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
