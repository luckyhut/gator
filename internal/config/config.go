package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() Config {
	// open ~/.gatorconfig
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	filepath := home + "/" + configFileName
	data, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
	}

	// read into Config struct
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatal(err)
	}

	return config
}

func (conf Config) SetUser(name string) {
	conf.CurrentUserName = name
	write(conf)
}

func getConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	fullPath := filepath.Join(home, configFileName)
	return fullPath, nil
}

func write(conf Config) error {
	fullPath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(conf)
	if err != nil {
		return err
	}

	return nil
}
