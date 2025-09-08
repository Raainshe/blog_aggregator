package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	DB_URL            string `json:"db_url"`
	Current_User_Name string `json:"current_user_name"`
}

const configFileName = ".gatorconfig.json"

func Read() (Config, error) {

	dir, err := os.UserHomeDir()
	if err != nil {
		return Config{}, fmt.Errorf("error getting user directory: %w", err)
	}
	dir += "/.gatorconfig.json"
	file, err := os.ReadFile(dir)
	if err != nil {
		return Config{}, fmt.Errorf("error getting opening file: %w", err)
	}

	var newConf Config

	err = json.Unmarshal(file, &newConf)
	if err != nil {
		return Config{}, fmt.Errorf("error unmarshalling data %w", err)
	}

	return newConf, nil
}

func (c *Config) SetUser(name string) error {
	c.Current_User_Name = name

	dir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	err = os.WriteFile(dir+"/"+configFileName, data, os.ModeAppend)
	if err != nil {
		return err
	}
	return nil
}
