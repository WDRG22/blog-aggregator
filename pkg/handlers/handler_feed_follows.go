package handlers

import (
	"net/http"
        "time"
        "github.com/google/uuid"
        "github.com/wdrg22/blog-aggregator/internal/database"
        "github.com/wdrg22/blog-aggregator/pkg/middleware"
)


// Handles "POST /feed_follows"
// Creates new feed_follow record for logged-in user
func (cfg *ApiConfig) HandlerCreateFeedFollow(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	user, ok := r.Context().Value(middleware.UserContextKey).(database.User)
	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Get feed_id from form data. In UI, this will be hidden input
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	feedIDStr := r.PostFormValue("feed_id")
	feedID, err := uuid.Parse(feedIDStr)
	if err != nil {
		http.Error(w, "Invalid Feed ID", http.StatusBadRequest)
		return
	}

	// Create feed_follow record in db
	feedFollowParams := database.CreateFeedFollowParams{
		ID:		uuid.New(),
		CreatedAt:	time.Now(),
		UpdatedAt:	time.Now(),
		UserID:		user.ID,
		FeedID:		feedID,		
	}
	feedFollow, err := cfg.DB.CreateFeedFollow(r.Context(), feedFollowParams)
	if err != nil {
		http.Error(w, "Failed to follow feed", http.StatusInternalServerError)
		return
	}

	// Pass new feedFollow record to template
	cfg.renderPage(w, http.StatusOK, "unfollow_button.html", feedFollow)
}

// Handles "DELETE /feed_follows/{feedFollowID}"
// Deletes a feed_follow record for the logged-in user
func (cfg *ApiConfig) HandlerDeleteFeedFollow(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	user, ok := r.Context().Value(middleware.UserContextKey).(database.User)
	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Get feedFollowID from URL path
	feedFollowIDStr := r.PathValue("feedFollowID")
	feedFollowID, err := uuid.Parse(feedFollowIDStr)
	if err != nil {
		http.Error(w, "Invalid Feed Follow ID", http.StatusBadRequest)
		return
	}

	// Delete record from db
	deleteFeedFollowParams := database.DeleteFeedFollowParams{
		ID:	feedFollowID,
		UserID:	user.ID,
	}
	err = cfg.DB.DeleteFeedFollow(r.Context(), deleteFeedFollowParams)
	if err != nil {
		http.Error(w, "Failed to unfollow feed", http.StatusInternalServerError)
		return
	}

	// This endpoint is called via HTMX'S hx-delete. 
	// Since the button will be swapped out by a different mechanism, 
	// (e.g. swapping a whole list item), we can just return a 200 ok status with no content
	w.WriteHeader(http.StatusOK)
}
