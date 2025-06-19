
package main

import _ "github.com/lib/pq"
import (
	"fmt"
	"log"
	"database/sql"
	"net/http"
	"io/fs"
	"github.com/joho/godotenv"
	"github.com/wdrg22/blog-aggregator/internal/config"
	"github.com/wdrg22/blog-aggregator/internal/worker"
	"github.com/wdrg22/blog-aggregator/internal/database"
	"github.com/wdrg22/blog-aggregator/pkg/handlers"
	"github.com/wdrg22/blog-aggregator/pkg/middleware"
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
	appWorker := worker.NewWorker(dbQueries)
	go appWorker.Start(cfg.WorkerInterval, cfg.WorkerCount)

	// -- TEMPLATE CACHING --
	templateCache, err := ui.NewTemplateCache()
	if err != nil {
		log.Fatalf("Failed to build template cache: %v", err)
	}

	// -- DEPENDENCY INJECTION SETUP --
	// Pass template cache to handlers
	apiCfg := &handlers.ApiConfig{
		DB:		dbQueries,
		Templates:	templateCache,
	}

	// -- ROUTER SETUP -- 
	mux := http.NewServeMux()

	// Create new filesystem from "static" subdirectory of embedded FS
	staticFS, err := fs.Sub(ui.StaticFS, "static")
	if err != nil {
		log.Fatal(err)
	}

	// Create file server using this new sub-filesystem
	fileServer := http.FileServer(http.FS(staticFS))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	// -- Register public routes -- 
	mux.HandleFunc("GET /login", apiCfg.HandlerLoginPage)
	mux.HandleFunc("GET /register", apiCfg.HandlerRegisterPage)
	mux.HandleFunc("POST /users/register", apiCfg.HandlerRegister)
	mux.HandleFunc("POST /users/login", apiCfg.HandlerLogin)

	// -- Register private routes -- 
	mux.HandleFunc("GET /", middleware.RequireAuth(apiCfg.HandlerDashboard))
	mux.HandleFunc("POST /users/logout", middleware.RequireAuth(apiCfg.HandlerLogout))
	mux.HandleFunc("POST /feeds", middleware.RequireAuth(apiCfg.HandlerCreateFeed))
	mux.HandleFunc("POST /feed_follows", middleware.RequireAuth(apiCfg.HandlerCreateFeed))
	mux.HandleFunc("DELETE /feed_follows/{feedFollowID}", middleware.RequireAuth(apiCfg.HandlerCreateFeed))

	// -- WRAP ROUTER WITH GLOBAL MIDDLEWARE
	// PopulateUserContext runs on every request before the mux routes it
	// This makes user info available to all handlers if cookie is present
	var handler http.Handler = middleware.PopulateUserContext(mux, dbQueries)

	// -- START SERVER --
	server := &http.Server{
		Addr: fmt.Sprintf(":%s", cfg.Port),
		Handler: handler,
	}

	log.Printf("Starting web server on http://localhost:%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
