package command

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/luckyhut/gator/database"
	"github.com/luckyhut/gator/rss"
	"html"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func HandlerAgg(s *State, cmd Command) error {
	if len(cmd.Args) < 1 {
		return errors.New("Not enough arguments")
	}
	duration, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return errors.New("Error parsing duration")
	}

	fmt.Println("Collecting feeds every", duration)
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for ; ; <-ticker.C {
		err := scrapeFeeds(s)
		if err != nil {
			return err
		}
	}
}

func scrapeFeeds(s *State) error {
	ctx := context.Background()
	nextFeed, err := s.Db.GetNextFeedToFetch(ctx)
	if err != nil {
		return errors.New("Error getting next feed from database")
	}

	s.Db.MarkFeedFetched(ctx, nextFeed.ID)

	feed, err := fetchFeed(ctx, nextFeed.Url.String)
	if err != nil {
		return err
	}

	for _, item := range feed.Channel.Item {
		params := createPostParams(&item, &nextFeed)
		err = s.Db.CreatePost(ctx, *params)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				continue
			}
			fmt.Println("Error type: ", reflect.TypeOf(err))
			return errors.New("Error adding post to database")
		}
	}
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

	// xml.Unmarshal
	var feed rss.RSSFeed
	err = xml.Unmarshal(body, &feed)
	if err != nil {
		fmt.Println(feed)
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
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
	}
}

func createPostParams(item *rss.RSSItem, feed *database.Feed) *database.CreatePostParams {
	params := database.CreatePostParams{
		ID:          uuid.New(),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		Title:       sql.NullString{String: item.Title, Valid: true},
		Url:         item.Link,
		Description: sql.NullString{String: item.Description, Valid: true},
		PublishedAt: sql.NullString{String: item.PubDate, Valid: true},
		FeedID:      feed.ID,
	}
	return &params
}

func HandlerAddFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) < 2 {
		return errors.New("Must include a name and url with this command")
	}

	params := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      sql.NullString{String: cmd.Args[0], Valid: true},
		Url:       sql.NullString{String: cmd.Args[1], Valid: true},
		UserID:    uuid.NullUUID{UUID: user.ID, Valid: true},
	}

	s.Db.CreateFeed(context.Background(), params)

	feed_id, err := s.Db.GetFeed(context.Background(), sql.NullString{String: cmd.Args[1], Valid: true})
	if err != nil {
		return errors.New("Error getting feed from database")
	}

	feedFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feed_id,
	}

	s.Db.CreateFeedFollow(context.Background(), feedFollowParams)

	return nil
}

func HandlerFeeds(s *State, cmd Command) error {
	dbContext := context.Background()
	result, err := s.Db.GetAllFeeds(dbContext)
	if err != nil {
		return errors.New("Unable to get list of feeds from database")
	}
	fmt.Println("handlerFeeds result: ", result)

	return nil
}

func HandlerBrowse(s *State, cmd Command) error {
	numPosts := 0
	var err error
	if len(cmd.Args) == 0 {
		numPosts = 2
	} else {
		numPosts, err = strconv.Atoi(cmd.Args[0])
		if err != nil {
			return errors.New("Number of posts must be an integer")
		}
	}
	ctx := context.Background()
	userUuid, err := s.Db.GetUser(ctx, s.Config.CurrentUserName)
	if err != nil {
		return err
	}
	posts, err := s.Db.GetPostsForUser(ctx, userUuid.ID)
	if err != nil {
		return err
	}
	if len(posts) == 0 {
		return errors.New("No posts to display")
	}
	if len(posts) == 1 {
		fmt.Println("---------------------------------------------------")
		fmt.Printf("%s\n", posts[0].Title.String)
		fmt.Printf("%s\n", posts[0].Description.String)
		fmt.Printf("%s\n", posts[0].Url)
		fmt.Println("---------------------------------------------------")
	}
	for i := 0; i < numPosts; i++ {
		fmt.Println("---------------------------------------------------")
		fmt.Printf("%s\n", posts[i].Title.String)
		fmt.Printf("%s\n", posts[i].Description.String)
		fmt.Printf("%s\n", posts[i].Url)
		fmt.Println("---------------------------------------------------")
	}
	return nil
}

func HandlerFollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) == 0 {
		return errors.New("Must include arguments with command")
	}
	dbContext := context.Background()

	// id
	feedFollowUuid := uuid.New()

	// updated_at, created_at
	curTime := time.Now().UTC()

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

func HandlerUnfollow(s *State, cmd Command, user database.User) error {
	dbContext := context.Background()
	url := sql.NullString{String: cmd.Args[0], Valid: true}
	feed_id, err := s.Db.GetFeed(dbContext, url) // feed_id
	if err != nil {
		return errors.New("Error getting feed from database")
	}

	params := database.UnfollowParams{
		UserID: user.ID,
		FeedID: feed_id,
	}

	s.Db.Unfollow(dbContext, params)
	return nil
}

func HandlerFollowing(s *State, cmd Command, user database.User) error {
	// get user id
	dbContext := context.Background()

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

func MiddlewareLoggedIn(handler func(s *State, cmd Command, user database.User) error) func(*State, Command) error {
	return func(s *State, cmd Command) error {
		user, err := s.Db.GetUser(context.Background(), s.Config.CurrentUserName)
		if err != nil {
			return errors.New("User is not registered")
		}
		return handler(s, cmd, user)
	}
}
