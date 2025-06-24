package command

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/luckyhut/gator/database"
	"time"
)

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
	curTime := time.Now().UTC()

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
