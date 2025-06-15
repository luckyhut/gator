package internal

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
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
		return fmt.Errorf("User %s is not registered\n", cmd.Args[0])
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
		return errors.New("User is already registered")
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
		return errors.New("Unable to delete users from database")
	}
	fmt.Println("Users successfully deleted from database")

	err = s.Db.ResetFeeds(dbContext)
	if err != nil {
		return errors.New("Unable to delete feeds from database")
	}
	fmt.Println("Feeds successfully deleted from database")

	err = s.Db.ResetFeedFollow(dbContext)
	if err != nil {
		return errors.New("Unable to delete feed_follows from database")
	}
	fmt.Println("Feed_follows successfully deleted from database")
	return nil
}

func HandlerUsers(s *State, cmd Command) error {
	// needs a context
	dbContext := context.Background()

	users, err := s.Db.GetUsers(dbContext)
	if err != nil {
		return errors.New("Unable to get a list of users")
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

	result, err := fetchFeed(httpContext, cmd.Args[0])

	// exit if error
	if err != nil {
		errors.New("Error registering site")
	}

	fmt.Println(result)
	return nil
}

func fetchFeed(ctx context.Context, feedURL string) (*rss.RSSFeed, error) {
	// make an http request and client
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, errors.New("Unable to get a request")
	}
	client := &http.Client{}

	// set header, run request
	req.Header.Set("User-Agent", "gator")
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("Error running HTTP request")
	}
	defer resp.Body.Close()

	// use xml.Unmarshal to fit the response in a struct
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("Error reading response body")
	}

	// xml.Unmarshal (works the same as json.Unmarshal)
	var feed rss.RSSFeed
	err = xml.Unmarshal(body, &feed)
	fmt.Println(feed)
	if err != nil {
		return nil, errors.New("Error unmarshaling xml data")
	}

	unescapeHtml(&feed)

	return &feed, nil
}

func unescapeHtml(feed *rss.RSSFeed) {
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
	}
}

func HandlerAddFeed(s *State, cmd Command) error {
	if len(cmd.Args) < 2 {
		return errors.New("Must include a name and url with this command")
	}

	feedUuid := uuid.New()
	curTime := time.Now()
	dbContext := context.Background()
	name := sql.NullString{String: cmd.Args[0], Valid: true}
	url := sql.NullString{String: cmd.Args[1], Valid: true}
	user, err := s.Db.GetUuid(dbContext, s.Config.CurrentUserName)
	if err != nil {
		return errors.New("Error connecting to database")
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

	feed_id, err := s.Db.GetFeed(dbContext, url)
	if err != nil {
		return errors.New("Error getting feed from database")
	}

	feedFollowsUuid := uuid.New()
	feedFollowParams := database.CreateFeedFollowParams{
		ID:        feedFollowsUuid,
		CreatedAt: curTime,
		UpdatedAt: curTime,
		UserID:    user.ID,
		FeedID:    feed_id,
	}

	s.Db.CreateFeedFollow(dbContext, feedFollowParams)

	return nil
}

func HandlerFeeds(s *State, cmd Command) error {
	dbContext := context.Background()
	result, err := s.Db.GetAllFeeds(dbContext)
	if err != nil {
		return errors.New("Unable to get list of feeds from database")
	}
	fmt.Println(result)

	return nil
}

func HandlerFollow(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("Must include arguments with command")
	}
	dbContext := context.Background()

	// id
	feedFollowUuid := uuid.New()

	// updated_at, created_at
	curTime := time.Now()

	// user_id
	user, err := s.Db.GetUser(dbContext, s.Config.CurrentUserName)
	if err != nil {
		return errors.New("Error getting user from database")
	}
	//	userIdNull := uuid.NullUUID{UUID: user.ID, Valid: true}

	// feed_id
	url := sql.NullString{String: cmd.Args[0], Valid: true}
	feed_id, err := s.Db.GetFeed(dbContext, url) // feed_id
	if err != nil {
		return errors.New("Error getting feed from database")
	}

	params := database.CreateFeedFollowParams{
		ID:        feedFollowUuid,
		CreatedAt: curTime,
		UpdatedAt: curTime,
		UserID:    user.ID,
		FeedID:    feed_id,
	}

	err = s.Db.CreateFeedFollow(dbContext, params)
	if err != nil {
		return errors.New("Could not create FeedFollow record")
	}

	return nil
}

func HandlerFollowing(s *State, cmd Command) error {
	// get user id
	dbContext := context.Background()
	user, err := s.Db.GetUser(dbContext, s.Config.CurrentUserName)
	if err != nil {
		return errors.New("Could not get user")
	}

	// get list of follows
	result, err := s.Db.GetFeedFollowsForUser(dbContext, user.ID)
	if err != nil {
		return errors.New("Unable to get list of feeds from database")
	}

	if len(result) == 0 {
		return nil
	}

	for i := 0; i < len(result); i++ {
		fmt.Print("* ", result[i].FeedName.String, "\n")
	}

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
