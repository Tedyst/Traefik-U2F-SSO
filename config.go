package main

import (
	"github.com/tkanos/gonfig"
)

// Configuration is the main config
type Configuration struct {
	Port                int
	RegistrationAllowed bool
	RegistrationToken   string
}

func initConfig() error {
	err := gonfig.GetConf("config.json", &Config)
	if err != nil {
		return err
	}
	return nil
}
