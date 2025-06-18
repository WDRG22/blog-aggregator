
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
	"github.com/wdrg22/blog-aggregator/internal/config"
	"github.com/wdrg22/blog-aggregator/internal/database"
	"github.com/wdrg22/blog-aggregator/pkg/handlers"
	"github.com/lib/pq"
)

//go:embed all:../../ui/templates
var templateFiles embed.FS

//go:embed all:../../ui/static
var staticFiles embed.FS


func main() {
	// -- CONFIGURATION --
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if unspecified
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable"
		log.Println("DATABASE_URL not set, using default local connection")
	}
	
	// -- DATABASE CONNECTION --
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	dbQueries := database.New(db)

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
		Addr: fmt.Sprintf(":%s", port),
		Handler: handler,
	}

	log.Printf("Starting web server on http://localhost:%s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
