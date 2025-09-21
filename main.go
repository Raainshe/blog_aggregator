package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/raainshe/blog_aggregator/internal/config"
	"github.com/raainshe/blog_aggregator/internal/database"
	"github.com/raainshe/blog_aggregator/internal/rss"

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
	cliCommands.register("reset", handleReset)
	cliCommands.register("users", handleUsers)
	cliCommands.register("agg", handleAgg)
	cliCommands.register("addfeed", handleAddfeed)
	cliCommands.register("feeds", handleFeeds)
}

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("usage: gator <command> <args[]>")
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

func handleFeeds(s *state, cmd command) error {

	if len(cmd.args) != 0 {
		return fmt.Errorf("this command takes no arguments")
	}

	allFeeds, err := s.dbQueries.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get feed: %w", err)
	}

	var fmtFeed []string
	for i, feed := range allFeeds {
		user, err := s.dbQueries.GetUserbyID(context.Background(), feed.UserID)
		if err != nil {
			fmtFeed = append(fmtFeed, strconv.Itoa(i)+": N/A "+feed.Name+" "+feed.Url)
		} else {
			fmtFeed = append(fmtFeed, strconv.Itoa(i)+": "+user.Name+" "+feed.Name+feed.Url)
		}
	}

	for _, feed := range fmtFeed {
		println(feed)
	}
	return nil
}

func handleAgg(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("you have too many arguments")
	}
	_, err := rss.FetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}
	return nil
}

func handleUsers(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("you have too many arguments")
	}

	users, err := s.dbQueries.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get users from database: %w", err)
	}
	if len(users) == 0 {
		fmt.Println("There are currently no users in the databse. Use 'register <username>' to create a new user")
	}

	for i := len(users) - 1; i >= 0; i-- {
		if users[i].Name == s.cfg.Current_User_Name {
			fmt.Println("-", users[i].Name, "(current)")
		} else {
			fmt.Println("-", users[i].Name)
		}
	}
	return nil
}

func handleReset(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("you have too many arguments")
	}
	err := s.dbQueries.DeleteAllUsers(context.Background())
	if err != nil {
		return fmt.Errorf("failed to delete all users: %w", err)
	}
	err = s.dbQueries.DeleteAllFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("failed to delete all feeds: %w", err)
	}
	fmt.Println("Deleted all data from databse")
	return nil
}

func handleAddfeed(s *state, cmd command) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("the add feed commands expects two arguments. Usage: addfeed <name> <url>")
	}

	currentUser, err := s.dbQueries.GetUser(context.Background(), s.cfg.Current_User_Name)
	if err != nil {
		return fmt.Errorf("could not find user in database: %w", err)
	}
	newFeed := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
		Url:       cmd.args[1],
		UserID:    currentUser.ID,
	}
	feed, err := s.dbQueries.CreateFeed(context.Background(), newFeed)
	if err != nil {
		return err
	}
	fmt.Println("Succesfully created feed:", feed.Name, "at", feed.CreatedAt)
	fmt.Println(feed)
	return nil
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
