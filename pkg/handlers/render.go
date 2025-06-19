package handlers

import (
	"bytes"
	"log"
	"net/http"
	"html/template"
	"github.com/wdrg22/blog-aggregator/internal/database"
)


// Contains dependencies for all handlers
type ApiConfig struct {
	DB		*database.Queries
	Templates	map[string]*template.Template
}

// Handles rendering full HTML pages with base layout for HTMX responses. 
// Looks up template from cache and executes it with 
// provided data, and writes the response
func (cfg *ApiConfig) renderPage(w http.ResponseWriter, status int, page string, data any) {
	// Retrieve appropriate template set from cache by page name
	ts, ok := cfg.Templates[page]
	if !ok {
		// Template missing
		log.Printf("FATAL: Template %s not found in cache", page)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Create buffer to perform 'test render' to catch any 
	// execution errors before writing to http.ResponseWriter
	buf := new(bytes.Buffer)
	err := ts.ExecuteTemplate(buf,"base", data)
	if err != nil {
		log.Printf("Render Error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// If template executes w/o error, write http status code
	// and write contents of buffer to http.ResponseWriter
	w.WriteHeader(status)
	_, err = buf.WriteTo(w)
	if err != nil {
		log.Printf("Write Error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Handles rendering HTML partials for HTMX responses
func (cfg *ApiConfig) renderPartial(w http.ResponseWriter, status int, templateName string, data any) {
	// All partials are included in every page's template set.
	// We can just grab one template set to get all partials
	ts, ok := cfg.Templates["dashboard.html"]
	if !ok {
		log.Printf("FATAL: Template dashboard.html not found in cache")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Execute the specific template name passed in, not the "base" template
	err := ts.ExecuteTemplate(w, templateName, data)
	if err != nil {
		log.Printf("Partial Render Error: %v", err)
	}
}
