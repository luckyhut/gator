package main

import _ "github.com/lib/pq"

import (
	"fmt"
	"github.com/luckyhut/gator/internal"
	"github.com/luckyhut/gator/internal/config"
	"os"
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

	// get command
	if len(os.Args) < 3 {
		fmt.Println("error: not enough arguments")
		os.Exit(1)
	}

	current_command := &internal.Command{
		Name: os.Args[1],
		Args: os.Args[2:],
	}
	fmt.Println(*current_command)
	commands.Run(state, *current_command)
}
