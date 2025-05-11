package main

import (
	"encoding/xml"
	"net/http"
	"fmt"
	"io"
	"context"
	"html"
)

type RSSFeed struct {
	Channel struct {
		Title		string		`xml:"title"`
		Link		string		`xml:"link"`
		Description	string		`xml:"description"`
		Item		[]RSSItem 	`xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title		string		`xml:"title"`
	Link		string		`xml:"link"`
	Description	string		`xml:"description"`
	PubDate		string		`xml:"pubDate"`
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
	for _, item := range rssFeed.Channel.Item {
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
	}

	return &rssFeed, nil
}
