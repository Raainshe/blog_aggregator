package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/raainshe/blog_aggregator/internal/config"
)

type state struct {
	cfg *config.Config
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
		fmt.Printf("error: %w", err)
		os.Exit(1)
		return
	}
	var cmd command
	cmd.args = args[2:]
	cmd.name = args[1]
	err = cliCommands.run(&cliState, cmd)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

}

func handlerLogin(s *state, cmd command) error {

	if len(cmd.args) != 1 {
		return fmt.Errorf("the login command expects a single argument: gator <username>")
	}
	err := s.cfg.SetUser(cmd.args[0])
	if err != nil {
		return err
	}
	return nil
}

func (c *commands) run(s *state, cmd command) error {

	handler, exists := c.cmd[cmd.name]
	if !exists {
		return fmt.Errorf("commands does not exist")
	}
	return handler(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.mutex.Lock()
	c.cmd[name] = f
	c.mutex.Unlock()

}
