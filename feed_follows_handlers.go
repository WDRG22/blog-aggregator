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


func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Usage: %s <url>", cmd.name)
	}

	// Get feed record from db by url
	feedUrl := cmd.args[0]
	feedRecord, err := s.db.GetFeedByUrl(context.Background(), feedUrl)
	if err != nil {
		return err
	}

	// Create new feed-follow record for current user in db
	feedFollowParams := database.CreateFeedFollowParams{
		ID:		uuid.New(),
		CreatedAt:	time.Now(),
		UpdatedAt:	time.Now(),
		UserID:		user.ID,
		FeedID:		feedRecord.ID,		
	}
	feedFollowRecord, err := s.db.CreateFeedFollow(context.Background(), feedFollowParams)
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
	fmt.Printf("%s is now following %s\n", feedFollowRecord.UserName, feedFollowRecord.FeedName)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	if len(cmd.args) > 0 {
		return fmt.Errorf("Usage: %s", cmd.name)
	}

	// Get feed-follow records by user ID
	feedFollowRecords, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID) 
	if err != nil {
		return fmt.Errorf("Error retrieving feed-follow records: %w", err)
	}

	// Print name of feeds being followed 
	for _, record := range feedFollowRecords {
		fmt.Println(record.FeedName)
	}
	return nil

}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Usage: %s <url>", cmd.name)
	}
	feedUrl := cmd.args[0]
	
	// Get Feed record from db
	feed, err := s.db.GetFeedByUrl(context.Background(), feedUrl)
	if err != nil {
		return err
	}
	params := database.DeleteFeedFollowParams{user.ID, feed.ID}

	// Delete feed-follow record
	err = s.db.DeleteFeedFollow(context.Background(), params) 
	if err != nil {
		return fmt.Errorf("Error deleting feed-follow record from database: %w", err)
	}
	return nil
}
