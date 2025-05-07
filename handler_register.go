package main

import (
	"fmt"
	"context"
	"database"
	"time"
)

func handlerRegister(s *state, cmd command) error {
	// Ensure proper usage of register command
	if len(cmd.args) != 1 { 
		return fmt.Errorf("Usage: %s <name>", cmd.name)
	}
	name := cmd.args[0]

	// Create new user params
	userParams := s.db.CreateUserParams{
		ID: 		uuid.New(),
		CreatedAt: 	time.Now(),
		UpdatedAt: 	time.Now(),
		Name:		name 
	}

	// Create user in db
	user, err := s.db.CreateUser(context.Background(), userParams)

	// Catch errors
	if err != nil {
		// If user already exists, return 1
		pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code.Name() == "unique_violation" {
				return 1
			}
		}
		// Else return err
		return fmt.Errorf("Error registering new user: %s", err)
	}

	// Set current user to new user in config
	s.cfg.SetUser(name)

	// Print new user data
	fmt.Println("User was created")
	fmt.Println(userParams)
	return nil
}
