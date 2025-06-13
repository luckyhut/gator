package internal

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"time"

	"database/sql"
	"github.com/google/uuid"
	"github.com/luckyhut/gator/internal/config"
	"github.com/luckyhut/gator/internal/database"
	"github.com/luckyhut/gator/internal/rss"
)

type Command struct {
	Name string
	Args []string
}

type State struct {
	Config *config.Config
	Db     *database.Queries
}

func HandlerLogin(s *State, cmd Command) error {
	// check that a user is included
	if len(cmd.Args) == 0 {
		return errors.New("Must include arguments with command")
	}

	// make sure user is in the db
	dbContext := context.Background()
	_, err := s.Db.GetUser(dbContext, cmd.Args[0])
	if err != nil {
		fmt.Printf("User %s is not registered\n", cmd.Args[0])
		os.Exit(1)
	}

	// set the user in config file
	err = s.Config.SetUser(cmd.Args[0])
	if err != nil {
		return err
	}
	fmt.Printf("User %s set.\n", cmd.Args[0])
	return nil
}

func HandlerRegister(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("Must include arguments with command")
	}

	// make params for new user
	dbContext := context.Background()
	userUuid := uuid.New()
	curTime := time.Now()

	// check to see if user is registered
	_, err := s.Db.GetUser(dbContext, cmd.Args[0])
	if err == nil {
		fmt.Println("User is already registered")
		os.Exit(1)
	}

	// create user to pass to database
	params := database.CreateUserParams{
		ID:        userUuid,
		CreatedAt: curTime,
		UpdatedAt: curTime,
		Name:      cmd.Args[0],
	}

	// add user to database
	s.Db.CreateUser(dbContext, params)
	fmt.Printf("User %s was created.\n", cmd.Args[0])
	err = HandlerLogin(s, cmd)
	if err != nil {
		return err
	}
	return nil
}

func HandlerReset(s *State, cmd Command) error {
	// needs a context
	dbContext := context.Background()

	err := s.Db.ResetUsers(dbContext)
	if err != nil {
		fmt.Println("Unable to delete users from database")
		os.Exit(1)
	}
	fmt.Println("Users successfully deleted from database")
	return nil
}

func HandlerUsers(s *State, cmd Command) error {
	// needs a context
	dbContext := context.Background()

	users, err := s.Db.GetUsers(dbContext)
	if err != nil {
		fmt.Println("Unable to get a list of users")
		os.Exit(1)
	}
	printUsers(s, users)
	return nil
}

func printUsers(s *State, users []string) {
	for i := range users {
		fmt.Printf("* %s", users[i])
		if s.Config.CurrentUserName == users[i] {
			fmt.Printf(" (current)")
		}
		fmt.Printf("\n")
	}
}

func HandlerAgg(s *State, cmd Command) error {
	// create a context and call helper
	httpContext := context.Background()
	// url := cmd.Args[0]
	result, err := fetchFeed(httpContext, "https://www.wagslane.dev/index.xml")

	// exit if error
	if err != nil {
		fmt.Println("Error registering site")
		os.Exit(1)
	}

	fmt.Println(result)
	return nil
}

func fetchFeed(ctx context.Context, feedURL string) (*rss.RSSFeed, error) {
	// make an http request and client
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		fmt.Println("Unable to get a request")
		os.Exit(1)
	}
	client := &http.Client{}

	// set header, run request
	req.Header.Set("User-Agent", "gator")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error running HTTP request")
		os.Exit(1)
	}
	defer resp.Body.Close()

	// use xml.Unmarshal to fit the response in a struct
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body")
		os.Exit(1)
	}

	// xml.Unmarshal (works the same as json.Unmarshal)
	var feed rss.RSSFeed
	err = xml.Unmarshal(body, &feed)
	fmt.Println(feed)
	if err != nil {
		fmt.Println("Error unmarshaling xml data")
		os.Exit(1)
	}

	_ = unescapeHtml(&feed)

	return &feed, nil
}

func unescapeHtml(feed *rss.RSSFeed) error {
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
	}
	return nil
}

func HandlerAddFeed(s *State, cmd Command) error {
	if len(cmd.Args) < 2 {
		fmt.Println("Must include a name and url with this command")
		os.Exit(1)
	}

	feedUuid := uuid.New()
	curTime := time.Now()
	dbContext := context.Background()
	name := sql.NullString{String: cmd.Args[0], Valid: true}
	url := sql.NullString{String: cmd.Args[1], Valid: true}
	user, err := s.Db.GetUuid(dbContext, s.Config.CurrentUserName)
	if err != nil {
		fmt.Println("Error connecting to database")
		os.Exit(1)
	}
	userUuid := uuid.NullUUID{UUID: user.ID, Valid: true}

	params := database.CreateFeedParams{
		ID:        feedUuid,
		CreatedAt: curTime,
		UpdatedAt: curTime,
		Name:      name,
		Url:       url,
		UserID:    userUuid,
	}

	// run createfeed
	s.Db.CreateFeed(dbContext, params)

	fmt.Println("Params: ---------- \n", params)

	return nil
}

type Commands struct {
	Commands_list map[string]func(*State, Command) error
}

func (c *Commands) Run(s *State, cmd Command) error {
	f, exists := c.Commands_list[cmd.Name]
	if !exists {
		return errors.New("command not found")
	}
	return f(s, cmd)
}

func (c *Commands) Register(name string, f func(*State, Command) error) {
	c.Commands_list[name] = f
}
