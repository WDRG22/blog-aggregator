package handlers

import (
        "time"
	"net/http"
	"github.com/lib/pq"
        "github.com/google/uuid"
        "github.com/wdrg22/blog-aggregator/internal/database"
)



// Renders login page
func (cfg *ApiConfig) HandlerLoginPage(w http.ResponseWriter, r *http.Request) {
	cfg.renderPage(w, http.StatusOK, "login.html", nil)
}

// Renders register page
func (cfg *ApiConfig) HandlerRegisterPage(w http.ResponseWriter, r *http.Request) {
	cfg.renderPage(w, http.StatusOK, "register.html", nil)
}

// Responds to registration POST request by creating new user in db and redirecting browser to homepage or displaying an error
func (cfg *ApiConfig) HandlerRegister(w http.ResponseWriter, r *http.Request) {
	// Get input data from form	
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
        name := r.PostFormValue("name")
	if name == "" {
		http.Error(w, "Name cannot be empty", http.StatusBadRequest)
		return
	}

        // Create new user in database
        userParams := database.CreateUserParams{
                ID:             uuid.New(),
                CreatedAt:      time.Now(),
                UpdatedAt:      time.Now(),
                Name:           name,
        }
        user, err := cfg.DB.CreateUser(r.Context(), userParams)
        if err != nil {
                if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
			http.Error(w, "A user with that name already exists", http.StatusConflict) // 409 Conflict
			return
                }
		http.Error(w, "Error registering new user", http.StatusInternalServerError)
		return
        }

	// Set a session cookie and redirect
	http.SetCookie(w, &http.Cookie{
		Name:		"session_user",
		Value:		user.Name,
		Path:		"/",
		Expires:	time.Now().Add(72 * time.Hour),
		HttpOnly:	true,
		SameSite:	http.SameSiteLaxMode,
	})

	// Tell browser to redirect to homepage via htmx
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}

func (cfg *ApiConfig) HandlerLogin(w http.ResponseWriter, r *http.Request) {
	// Get input data from form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	name := r.PostFormValue("name")

	// Check if user exists in db
	if _, err := cfg.DB.GetUser(r.Context(), name); err != nil {
		http.Error(w, "Invalid username", http.StatusUnauthorized)
		return
	}

	// Set session cookie and redirect
	http.SetCookie(w, &http.Cookie{
		Name:		"session_user",
		Value:		name,
		Path:		"/",
		Expires:	time.Now().Add(72 * time.Hour),
		HttpOnly:	true,
		SameSite:	http.SameSiteLaxMode,
	})

	// Tell browser to redirect to homepage via htmx
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}

func (cfg *ApiConfig) HandlerLogout(w http.ResponseWriter, r *http.Request) {
	// To 'delete' cookie, we set it with same name but expired date in the past
	http.SetCookie(w, &http.Cookie{
		Name:		"session_user",
		Value:		"",
		Path:		"/",
		Expires:	time.Unix(0,0),
		HttpOnly:	true,
		SameSite:	http.SameSiteLaxMode,
	})

	// Tell browser to redirect to login page via htmx
	w.Header().Set("HX-Redirect", "/login")
	w.WriteHeader(http.StatusOK)
}
