package ui

import (
	"embed"
	"io/fs"
	"html/template"
	"path/filepath"
)

// This directive embeds the entire 'templates' directory.
//go:embed templates
var TemplateFS embed.FS

// This directive embeds the entire 'static' directory.
//go:embed static
var StaticFS embed.FS

// Creates cache of parsed templates. Occurs once at startup
// Each "view" template is parsed along with the base layout and any partials
func NewTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	// Get all "view" files
	pages, err := fs.Glob(TemplateFS, "templates/views/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		// Get filename to use as map key
		name := filepath.Base(page)

		// Create new template set for this page
		ts, err := template.New(name).ParseFS(TemplateFS,
		"templates/layouts/base.html",
		"templates/partials/feed_list.html",
		"templates/partials/no_feeds.html",
		)
		if err != nil {
			return nil, err
		}

		// Now parse specific view file into the set
		ts, err = ts.ParseFS(TemplateFS, page)
		if err != nil {
			return nil, err 
		}

		// Add full parsed template set to cache
		cache[name] = ts
	}
	return cache, nil
}
