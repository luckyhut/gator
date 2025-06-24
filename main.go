package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // required for sql interaction
	"github.com/luckyhut/gator/command"
	"github.com/luckyhut/gator/config"
	"github.com/luckyhut/gator/database"
	"log"
	"os"
)

func main() {
	// read config file
	conf := config.Read()

	// initialize state and commands
	state := &command.State{
		Config: &conf,
	}
	commands := &command.Commands{
		Commands_list: make(map[string]func(*command.State, command.Command) error),
	}
	commands.Register("login", command.HandlerLogin)
	commands.Register("register", command.HandlerRegister)
	commands.Register("reset", command.HandlerReset)
	commands.Register("users", command.HandlerUsers)
	commands.Register("agg", command.HandlerAgg)
	commands.Register("feeds", command.HandlerFeeds)
	commands.Register("browse", command.HandlerBrowse)
	commands.Register("addfeed", command.MiddlewareLoggedIn(command.HandlerAddFeed))
	commands.Register("follow", command.MiddlewareLoggedIn(command.HandlerFollow))
	commands.Register("unfollow", command.MiddlewareLoggedIn(command.HandlerUnfollow))
	commands.Register("following", command.MiddlewareLoggedIn(command.HandlerFollowing))

	// open connection to database
	db, err := sql.Open("postgres", state.Config.DbUrl)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)
	state.Db = dbQueries

	// get command
	if len(os.Args) < 2 {
		fmt.Println("Error, not enough arguments")
		os.Exit(1)
	}

	current_command := &command.Command{
		Name: os.Args[1],
		Args: os.Args[2:],
	}

	err = commands.Run(state, *current_command)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
