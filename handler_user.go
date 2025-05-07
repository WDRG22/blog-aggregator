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


func handlerLogin(s *state, cmd command) error {
	// Ensure proper usage of login command
	if len(cmd.args) != 1 {
		return fmt.Errorf("Usage: %s <name>", cmd.name)
	}
	userName := cmd.args[0]

	// Check if user in database
	_, err := s.db.GetUser(context.Background(), userName)
	if err != nil {
		return fmt.Errorf("User does not exist in database: %w", err)
	}

	// Set user in config
	err = s.cfg.SetUser(userName)
	if err != nil {
		return fmt.Errorf("Failed to set current user: %w\n", err) 
	}

	fmt.Printf("User has been set to: %s\n", userName)
	return nil
}


func handlerRegister(s *state, cmd command) error {
        // Ensure proper usage of register command
        if len(cmd.args) != 1 {
                return fmt.Errorf("Usage: %w <name>", cmd.name)
        }
        name := cmd.args[0]

        // Create new user params
        userParams := database.CreateUserParams{
                ID:             uuid.New(),
                CreatedAt:      time.Now(),
                UpdatedAt:      time.Now(),
                Name:           name,
        }

        // Create user in db
        user, err := s.db.CreateUser(context.Background(), userParams)

        // Catch errors
        if err != nil {
                // If user already exists, return 1
                if pqErr, ok := err.(*pq.Error); ok {
                        if pqErr.Code.Name() == "unique_violation" {
                                os.Exit(1)
                        }
                }
                // Else return err
                return fmt.Errorf("Error registering new user: %w", err)
        }

        // Set current user to new user in config
        err = s.cfg.SetUser(name)
        if err != nil {
                fmt.Errorf("Error updating config with new user: %w", err)
        }

        // Print new user data
        fmt.Println("User was created")
        fmt.Println(user)
        return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Error deleting all users: %w", err)
	}
	fmt.Println("Successfully deleted all users")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil{
		return fmt.Errorf("Error retrieving users: %w", err)
	}
	for _, user := range users {
		fmt.Println(user)
	}
	return nil
}
