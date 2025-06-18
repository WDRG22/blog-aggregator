
package main

import _ "github.com/lib/pq"
import (
	"fmt"
	"os"
	"log"
	"database/sql"
	"net/http"
	"embed"
	"html/template"
	"github.com/lib/pq"
	"github.com/joho/godotenv"
	"github.com/wdrg22/blog-aggregator/internal/config"
	"github.com/wdrg22/blog-aggregator/internal/database"
	"github.com/wdrg22/blog-aggregator/pkg/handlers"
	"github.com/wdrg22/blog-aggregator/ui"
)


func main() {
	// -- CONFIGURATION --
	err := godotenv.Load()
	if err != nil {
		log.Println("Note: .env file not found, reading from environment")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// -- DATABASE CONNECTION --
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	dbQueries := database.New(db)

	// -- START BACKGROUND WORKER
	go worker.Start(dbQueries, cfg.WorkerCount, cfg.WorkerInterval)

	// -- TEMPLATE PARSING --
	templates, err := template.ParseFS(templateFiles, "templates/**/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	// -- DEPENDENCY INJECTION SETUP --
	// apiCfg holds all dependencies that handlers need
	apiCfg := &handlers.ApiConfig{
		DB:		dbQueries,
		Templates:	templates,
		AppConfig: 	cfg,
	}

	// -- ROUTER SETUP -- 
	mux := http.NewServeMux()

	// Create file server for static assets from embedded filesystem
	// Call http.FS to convert embed.FS to http.FileSystem
	staticFS, _ := http.FS(staticFiles)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(staticFS)))

	// Register application routes
	mux.HandleFunc("GET /", apiCfg.HandlerHome)
	mux.HandleFunc("POST /users/register", apiCfg.HandlerRegisterUser)
	mux.HandlerFunc("POST /users/login", apiCfg.HandlerLoginUser)
	mux.HandlerFunc("POST /feed", apiCfg.HandlerCreateFeed)
	mux.HandlerFunc("GET /feeds", apiCfg.HandlerGetFeeds)

	// -- START SERVER --
	// Wrap mux with middleware
	handler := handlers.MiddlewareLogging(mux)

	server := &http.Server{
		Addr: fmt.Sprintf(":%s", cfgPport),
		Handler: handler,
	}

	log.Printf("Starting web server on http://localhost:%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
