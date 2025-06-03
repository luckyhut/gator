package internal

import (
	"errors"
	"fmt"
	"github.com/luckyhut/gator/internal/config"
)

type Command struct {
	Name string
	Args []string
}

type State struct {
	Config *config.Config
}

func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("Must include arguments with command")
	}
	err := s.Config.SetUser(cmd.Args[0])
	if err != nil {
		return err
	}
	fmt.Printf("User %s set.\n", cmd.Args[0])
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
