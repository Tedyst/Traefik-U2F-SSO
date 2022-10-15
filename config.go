package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config is imported from config.yml
var Config Configuration

// Configuration is the main config
type Configuration struct {
	Port                int    `json:"port"`
	RegistrationAllowed bool   `json:"registrationAllowed"`
	RegistrationToken   string `json:"registrationToken"`
	URL                 string `json:"URL"`
	Debug               bool   `json:"debug"`
	Domain              string `json:"domain"`
	SqliteFile          string `json:"sqliteFile"`
}

func initConfig() error {
	content, err := os.ReadFile("config.json")
	if err != nil {
		return fmt.Errorf("error when opening file: %w", err)
	}
	var conf Configuration
	err = json.Unmarshal(content, &conf)
	if err != nil {
		return fmt.Errorf("error parsing config: %w", err)
	}
	Config = conf
	return nil
}
