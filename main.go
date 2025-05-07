
package main

import _ "github.com/lib/pq"
import (
	"fmt"
	"os"
	"log"
	"github.com/wdrg22/blog-aggregator/internal/config"
)

dbURL = "postgres://postgres:postgres@localhost:5432/gator"

type state struct {
	cfg *config.Config
	db *database.Queries
}

func main() {
	
	// Read config file 
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("Error reading config: ", err)
		return
	}

	// Open db connection
	db, err := sql.Open("postgres", dbURL)
	dbQueries := database.New(db)

	// Store config and db queries in program state
	programState := &state{
		cfg: &cfg,
		db: &dbQueries
	}

	// Register cli commands
	cmds := commands{
		registeredCommands: make(map[string]func(*state, command) error),
	}
	cmds.register("login", handlerLogin)

	if len(os.Args) < 2 {
		log.Fatal("Usage: cli <command> [args...]")
		return
	}
	
	cmdName := os.Args[1]
	cmdArgs := os.Args[2:]

	err = cmds.run(programState, command{name: cmdName, args: cmdArgs})
	if err != nil {
		log.Fatal(err)
	}
}
