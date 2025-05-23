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
                return fmt.Errorf("Usage: %s <name>", cmd.name)
        }
        name := cmd.args[0]

        // Create new user parameters
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
        err = s.cfg.SetUser(user.Name)
        if err != nil {
                fmt.Errorf("Error updating config with new user: %w", err)
        }

        // Print new user data
        fmt.Printf("User \"%s\" has been registered\n", user.Name)
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

        if len(users) == 0 {
                fmt.Println("No registered users")
                return nil
        }

        for _, user := range users {
                if user.Name == s.cfg.CurrentUserName{
                        fmt.Printf("* %s (current)\n", user.Name)
                } else {
                        fmt.Printf("* %s\n", user.Name)
                }
        }
        return nil
}
