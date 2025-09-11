package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/raainshe/blog_aggregator/internal/config"
	"github.com/raainshe/blog_aggregator/internal/database"

	_ "github.com/lib/pq"
)

type state struct {
	cfg       *config.Config
	dbQueries *database.Queries
}

type command struct {
	name string
	args []string
}

type commands struct {
	mutex sync.Mutex
	cmd   map[string]func(*state, command) error
}

var cliState state
var cliCommands commands

func init() {
	cliCommands.cmd = make(map[string]func(*state, command) error)
	cliCommands.register("login", handlerLogin)
	cliCommands.register("register", handlerRegister)
}

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("usage: gator <command> <?args[]?>")
		os.Exit(1)
		return
	}

	newcfg, err := config.Read()
	cliState.cfg = &newcfg
	if err != nil {
		fmt.Printf("error: %v", err)
		os.Exit(1)
		return
	}
	db, err := sql.Open("postgres", newcfg.DB_URL)
	if err != nil {
		fmt.Printf("error with sql open: %v", err)
	}
	dbQueries := database.New(db)
	cliState.dbQueries = dbQueries
	var cmd command
	cmd.args = args[2:]
	cmd.name = args[1]
	err = cliCommands.run(&cliState, cmd)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

}

func handlerRegister(s *state, cmd command) error {

	if len(cmd.args) != 1 {
		return fmt.Errorf("the register commands expects one argument. Usage: gator register <username>")
	}
	newUser := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
	}
	//check if user exists as well
	_, err := s.dbQueries.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		newUser, err := s.dbQueries.CreateUser(context.Background(), newUser)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
		fmt.Println("Successfully create user:", newUser.Name)
		return handlerLogin(s, cmd)
	}
	return fmt.Errorf("user %s already exists in db", cmd.args[0])
}

func handlerLogin(s *state, cmd command) error {

	if len(cmd.args) != 1 {
		return fmt.Errorf("the login command expects a single argument: gator <username>")
	}

	//check they exist in the database
	user, err := s.dbQueries.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("could not find user %s in the databse", cmd.args[0])
	}
	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return err
	}
	return nil
}

func (c *commands) run(s *state, cmd command) error {

	handler, exists := c.cmd[cmd.name]
	if !exists {
		return fmt.Errorf("command does not exist")
	}
	return handler(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.mutex.Lock()
	c.cmd[name] = f
	c.mutex.Unlock()

}
