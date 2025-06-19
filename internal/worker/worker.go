package worker 

import (
	"encoding/xml"
	"net/http"
	"fmt"
	"io"
	"context"
	"html"
	"time"
	"log"
	"strings"
	"database/sql"
	"github.com/lib/pq"
	"github.com/google/uuid"
	"github.com/wdrg22/blog-aggregator/internal/database"
)


var timeLayouts = []string{
	time.RFC1123Z,
	time.RFC3339,
	time.RFC1123,
	time.RFC822Z,
	time.RFC822,
	"2006-01-02T15:04:05-07:00", // A common alternative format
}

type Worker struct {
	DB *database.Queries
}

type RSSFeed struct {
	Channel struct {
		Title		string		`xml:"title"`
		Link		string		`xml:"link"`
		Description	string		`xml:"description"`
		Items		[]RSSItem 	`xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title		string		`xml:"title"`
	Link		string		`xml:"link"`
	Description	string		`xml:"description"`
	PubDate		string		`xml:"pubDate"`
} 


func NewWorker(db *database.Queries) *Worker {
	return &Worker{DB: db}
}

// Start ticking loop that fetches feeds
func (w *Worker) Start(interval time.Duration, concurrency int) {
	log.Printf("Starting worker: collecting feeds every %s", interval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	// Infinite loop in own goroutine
	for ; ; <- ticker.C {
		log.Println("Worker tick: searching for next feeds to fetch...")
		err := w.scrapeNext()
		if err != nil {
			log.Printf("Error scraping feed: %v", err)
		}
	}
}

// Fetches next due feed and processes its posts
func (w *Worker) scrapeNext() error {
	nextFeed, err := w.DB.GetNextFeedToFetch(context.Background())
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("No feeds due for fetching")
			return nil
		}
		return fmt.Errorf("Error getting next feed: %w", err)
	}

	log.Printf("Fetching feed: %s", nextFeed.Name)
	err = w.DB.MarkFeedFetched(context.Background(), nextFeed.ID)
	if err != nil {
		return fmt.Errorf("Error marking feed as fetched: %w", err)
	}

	rssFeed, err := FetchRSS(context.Background(), nextFeed.Url)
	if err != nil {
		return fmt.Errorf("Error fetching feed content: %w", err)
	}

	for _, item := range rssFeed.Channel.Items {
		pubDate, err := parseTime(item.PubDate)
		if err != nil {
			log.Printf("Could not parse publish date '%s' for post '%s': %v", item.PubDate, item.Title, err)
			continue
		}
		
		postParams := database.CreatePostParams{
			ID: 		uuid.New(),
			CreatedAt: 	time.Now().UTC(),
			UpdatedAt: 	time.Now().UTC(),
			Url: 		item.Link,
			PublishedAt: 	pubDate,
			FeedID: 	nextFeed.ID,
			Title: 		sql.NullString{String: item.Title, Valid: true},
			Description: sql.NullString{String: item.Description, Valid: true},
		}
		_, err = w.DB.CreatePost(context.Background(), postParams)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
				// Post already exists. Expected - not an error
				continue
			}
			log.Printf("Failed to create post '%s' in DB: %v", item.Title, err)
		}
	}
	log.Printf("Finished processing feed: %s. Found %d new posts.", nextFeed.Name, len(rssFeed.Channel.Items))
	return nil
}

func FetchRSS(ctx context.Context, feedURL string) (*RSSFeed, error) {
	if feedURL == "" {
		return nil, fmt.Errorf("Error, missing feed URL")
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("Error making request: %w", err)
	}
	req.Header.Set("User-Agent", "gator")

	// Create client
	client := &http.Client{Timeout: 10 * time.Second}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error sending request: %w", err)
	}
	
	// Read response body to byte slice
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %w", err)
	}

	// Unmarshal xml bytes into rssFeed struct
	var rssFeed RSSFeed
	err = xml.Unmarshal(body, &rssFeed)
	if err != nil {
		log.Printf("DEBUG: Failed to parse XML. Body received: %s", string(body))
		return nil, fmt.Errorf("Error unmarshalling data: %w", err)
	}

	// Decode escaped HTML entities in channel and item titles and descriptions
	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)
	for _, item := range rssFeed.Channel.Items {
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
	}

	return &rssFeed, nil
}

func parseTime(dateString string) (time.Time, error) {
	dateString = strings.TrimSpace(dateString)
	for _, layout := range timeLayouts {
		t, err := time.Parse(layout, dateString)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse time '%s' with any known layout", dateString)
}
