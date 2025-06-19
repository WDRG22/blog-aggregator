package handlers 

import (
        "time"
	"log"
	"net/http"
        "github.com/google/uuid"
        "github.com/wdrg22/blog-aggregator/internal/database"
        "github.com/wdrg22/blog-aggregator/internal/worker"
        "github.com/wdrg22/blog-aggregator/pkg/middleware"
)

// Main page for logged-in user
// Handles requests for "GET /"
func (cfg *ApiConfig) HandlerDashboard(w http.ResponseWriter, r *http.Request) {
	// Auth middleware provides user via request context
	user, ok := r.Context().Value(middleware.UserContextKey).(database.User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Fetch latest posts for this user
	userPostParams := database.GetPostsForUserParams{
		UserID: user.ID,
		Limit: 25,
	}
	posts, err := cfg.DB.GetPostsForUser(r.Context(), userPostParams)

	if err != nil {
		log.Println("ERROR: Failed to get posts for user")
		http.Error(w, "Failed to retrieve posts", http.StatusInternalServerError)
		return
	}

	// Fetch feeds followed by user
	feedFollows, err := cfg.DB.GetFeedFollowsForUser(r.Context(), user.ID)
	if err != nil {
		log.Printf("ERROR: Failed to retrieve followed feeds for '%s': %v",user.Name, err)
		http.Error(w, "Failed to retrieve followed feeds", http.StatusInternalServerError)
		return
	}

	// Bundle all necessary data into struct ot pass to template
	pageData := struct {
		User		database.User
		Posts		[]database.Post
		FeedFollows	[]database.GetFeedFollowsForUserRow
	}{
		User:		user,
		Posts:		posts,
		FeedFollows:	feedFollows,
	}

	// Render main dashboard template with user's data
	cfg.renderPage(w, http.StatusOK, "dashboard.html", pageData)
}

// Adds a new feed and creates the inital feed_follow record
// Hanldes "POST /feeds"
func (cfg *ApiConfig) HandlerCreateFeed(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	user, ok := r.Context().Value(middleware.UserContextKey).(database.User)
	if !ok {
		log.Printf("ERROR: Authentication failed")
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Parse form to get feed URL from <input name="url"> field
	if err := r.ParseForm(); err != nil {
		log.Printf("ERROR: Failed to parse form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	feedURL := r.PostFormValue("url")

	// Fetch feed's metadata to get its title
	rssFeed, err := worker.FetchRSS(r.Context(), feedURL)
	if err != nil {
		log.Printf("ERROR: Failed to fetch RSS feed '%s': %v", feedURL, err)
		http.Error(w, "Failed to fetch feed: " + err.Error(), http.StatusBadRequest)
		return
	}

	// Create feed record in database
	feedParams := database.CreateFeedParams{
		ID:		uuid.New(),
		CreatedAt:	time.Now().UTC(),
		UpdatedAt:	time.Now().UTC(),
		Name:		rssFeed.Channel.Title,
		Url:		feedURL,
		UserID:		user.ID,
	}
	feed, err := cfg.DB.CreateFeed(r.Context(), feedParams)
	if err != nil {
		log.Printf("ERROR: Failed to create feed in database: %v",user.Name, err)
		http.Error(w, "Failed to create feed in database", http.StatusInternalServerError)
		return
	}
	
	// Create feed_follow record for user
	feedFollowParams := database.CreateFeedFollowParams{
		ID:		uuid.New(),
		CreatedAt:	time.Now().UTC(),
		UpdatedAt:	time.Now().UTC(),
		UserID:		user.ID,
		FeedID:		feed.ID,
	}
	_, err = cfg.DB.CreateFeedFollow(r.Context(), feedFollowParams)
	if err != nil {
		log.Println("Failed to follow the new feed")
		http.Error(w, "Failed to follow the new feed", http.StatusInternalServerError)
		return
	}

	// Respond with HTML partial containing the updated list of followed feeds
	// HTMX will use this to dynamically update the feed list on the dashboard
	feedFollows, err := cfg.DB.GetFeedFollowsForUser(r.Context(), user.ID)
	if err != nil {
		log.Println("Could not retrieve updated feed list")
		http.Error(w, "Could not retrieve updated feed list", http.StatusInternalServerError)
		return
	}

	templateData := struct {
		FeedFollows []database.GetFeedFollowsForUserRow
	}{
		FeedFollows: feedFollows,
	}

	// Execute ONLY the partial template for the feed list
	cfg.renderPartial(w, http.StatusOK, "feed_list.html", templateData)  
}
