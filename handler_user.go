package main

import (
	"fmt"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Usage: %s <name>", cmd.name)
	}

	user := cmd.args[0]

	err := s.cfg.SetUser(user)
	if err != nil {
		return fmt.Errorf("Failed to set current user: %w\n", err) 
	}

	fmt.Printf("User has been set to: %s\n", user)
	return nil
}
