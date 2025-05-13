package main

import (
        "fmt"
        "context"
        "time"
        "os"
        "github.com/lib/pq"
        "github.com/google/uuid"
        "github.com/wdrg22/blog-aggregator/internal/database"
)


func handlerAgg(s *state, cmd command) error {
        rssFeed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
        if err != nil {
                return fmt.Errorf("Error fetching feed: %w", err)
        }
        fmt.Println(rssFeed)
        return nil
}

func handlerAddFeed(s *state, cmd command) error {
        // Ensure correct usage
        if len(cmd.args) != 2 {
                return fmt.Errorf("Usage: %s <name> <url>", cmd.name)
        }

        // Get current user from db
        currentUserName := s.cfg.CurrentUserName
        currentUserRecord, err := s.db.GetUser(context.Background(), currentUserName)
        if err != nil {
                return fmt.Errorf("Error getting current user from database: %w", err)
        }


        // Create new feed parameters
        feedName := cmd.args[0]
        feedURL := cmd.args[1]
        feedParams := database.CreateFeedParams{
                ID:             uuid.New(),
                CreatedAt:      time.Now(),
                UpdatedAt:      time.Now(),
                Name:           feedName,
                Url:            feedURL,
                UserID:         currentUserRecord.ID,
        }

        // Create feed record in database
        feedRecord, err := s.db.CreateFeed(context.Background(), feedParams)
        if err != nil {
                // If feed already exists, return 1
                if pqErr, ok := err.(*pq.Error); ok {
                        if pqErr.Code.Name() == "unique_violation" {
                                os.Exit(1)
                        }
                }
                // Else return err
                return fmt.Errorf("Error adding new feed: %w", err)
        }

	// Create new feed-follow record for current user and new feed
	followCmd := command {
		name: "follow",
		args: []string{feedURL},
	}
	err = handlerFollow(s, followCmd)
	if err != nil {
		return fmt.Errorf("Error running follow command within addfeed cmd: %w", err)
	}

        fmt.Printf("Successfully added new feed: %s\n", feedName)
	fmt.Printf("Name: %s\n", feedRecord.Name)
	fmt.Printf("URL: %s\n", feedRecord.Url)
        return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("Error retrieving feeds: %w", err)
	}

	if len(feeds) == 0 {
		fmt.Println("No feeds in database")
		return nil
	}

	for _, feed := range feeds {
		userRecord, err := s.db.GetUserById(context.Background(), feed.UserID)
		if err != nil {
			return fmt.Errorf("Error retrieving user record for feed \"%s\": %w", feed.Name, err)
		}
		userName := userRecord.Name
		fmt.Printf("Name: %s\n", feed.Name)
		fmt.Printf("URL: %s\n", feed.Url)
		fmt.Printf("Created By: %s\n\n", userName)
	}
	return nil
}
