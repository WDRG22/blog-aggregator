package main

import (
	"errors"
)

type command struct {
	name string
	args []string

}

type commands struct {
	registeredCommands map[string]func(*state, command) error
}


func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.registeredCommands[cmd.name]
	if !ok {
		return errors.New("Command not found")
	}
	return handler(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.registeredCommands[name] = f
}

