package worker 

import (
	"encoding/xml"
	"net/http"
	"fmt"
	"io"
	"context"
	"html"
	"time"
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

func parseTime(dateString string) (time.Time, error) {
	dateString = strings.TrimSpace(dateString)
	for _, layout := range timeLayouts{
		t, err := time.Parse(layout, dateString)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse tine '%s' with any known layout", dateString) 
}

func scrapeFeeds(s *state) error {
	nextFeedRecord, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("Error getting next feed: %w", err)
	}

	err = s.db.MarkFeedFetched(context.Background(), nextFeedRecord.ID)
	if err != nil {
		return fmt.Errorf("Error marking feed as fetched: %w", err)
	}

	feed, err := fetchFeed(context.Background(), nextFeedRecord.Url) 
	if err != nil {
		return fmt.Errorf("Error fetching feed: %w", err)
	}

	for _, item := range feed.Channel.Items {
		fmt.Printf("\nTitle: %s\n", item.Title)
		fmt.Printf("URL: %s\n", item.Link)
		// Parse PubDate strings across possible formats
		pubDate, err := parseTime(item.PubDate)
		if err != nil {
			return fmt.Errorf("Error parsing publish date: %w", err)
		}
		postParams := database.CreatePostParams{
			ID: uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Url: item.Link,
			PublishedAt: pubDate,
			FeedID: nextFeedRecord.ID,
			Title: sql.NullString{String: item.Title, Valid: true},
			Description: sql.NullString{String: item.Description, Valid: true},
		}
		_, err =
		s.db.CreatePost(context.Background(), postParams)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				if pqErr.Code.Name() == "unique_violation"{
					if pqErr.Constraint == "posts_url_key" {
						fmt.Printf("Post with URL '%s' already exists, skipping. \n", postParams.Url)
						continue
					}
				}
			}
			return fmt.Errorf("Error adding new post: %w", err)
		}
	}
	return nil
}
	
func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
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
	client := &http.Client{}

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
