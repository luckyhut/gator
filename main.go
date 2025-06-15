package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/luckyhut/gator/internal"
	"github.com/luckyhut/gator/internal/config"
	"github.com/luckyhut/gator/internal/database"
)

func main() {
	// read config file
	conf := config.Read()

	// initialize state and commands
	state := &internal.State{
		Config: &conf,
	}
	commands := &internal.Commands{
		Commands_list: make(map[string]func(*internal.State, internal.Command) error),
	}
	commands.Register("login", internal.HandlerLogin)
	commands.Register("register", internal.HandlerRegister)
	commands.Register("reset", internal.HandlerReset)
	commands.Register("users", internal.HandlerUsers)
	commands.Register("agg", internal.HandlerAgg)
	commands.Register("addfeed", internal.HandlerAddFeed)
	commands.Register("feeds", internal.HandlerFeeds)
	commands.Register("follow", internal.HandlerFollow)
	commands.Register("following", internal.HandlerFollowing)

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

	current_command := &internal.Command{
		Name: os.Args[1],
		Args: os.Args[2:],
	}

	err = commands.Run(state, *current_command)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
