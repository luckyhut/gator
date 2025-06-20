package command

import (
	"errors"
	"github.com/luckyhut/gator/config"
	"github.com/luckyhut/gator/database"
)

type Command struct {
	Name string
	Args []string
}

type State struct {
	Config *config.Config
	Db     *database.Queries
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
