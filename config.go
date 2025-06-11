// Copyright (c) 2024 Julian Müller (ChaoticByte)

package main

import (
	"errors"
)

type Config struct {
	ApiFetchInterval int `json:"api_fetch_interval"` // in seconds
	EnabledApiEndpoints []string `json:"enabled_api_endpoints"`
	PersistentDataFilePath string `json:"datafile"`
	LogLevel int `json:"loglevel"`
	Lists *[]NotifyList `json:"lists"`
	SmtpConfiguration SmtpSettings `json:"smtp"`
	Template MailTemplateConfig `json:"template"`
}

func NewConfig() Config {
	// Initial config
	c := Config{
		ApiFetchInterval: 60 * 10, // every 10 minutes,
		EnabledApiEndpoints: []string{"bay", "bund"},
		PersistentDataFilePath: "data.json",
		LogLevel: 2,
		Lists: &[]NotifyList{
			{ Name: "Example List",
			  Recipients: []string{},
			  Filter: []Filter{},},
		},
		SmtpConfiguration: SmtpSettings{
			From: "user@localhost",
			User: "user@localhost",
			Password: "change me :)",
			ServerHost: "127.0.0.1",
			ServerPort: 587},
		Template: MailTemplateConfig{
			SubjectTemplate: "",
			BodyTemplate: "",
		},
	}
	return c
}

func checkConfig(config Config) {
	if len(*config.Lists) < 1 {
		logger.error("Configuration is incomplete")
		panic(errors.New("no lists are configured"))
	}
	for _, l := range *config.Lists {
		if len(l.Filter) < 1 {
			logger.error("Configuration is incomplete")
			panic(errors.New("list " + l.Name + " has no filter defined - at least [{'any': true/false}] should be configured"))
		}
		for _, r := range l.Recipients {
			if !mailAddressIsValid(r) {
				logger.error("Configuration includes invalid data")
				panic(errors.New("'" + r + "' is not a valid e-mail address"))
			}
		}
	}
	if !mailAddressIsValid(config.SmtpConfiguration.From) {
		logger.error("Configuration includes invalid data")
		panic(errors.New("'" + config.SmtpConfiguration.From + "' is not a valid e-mail address"))
	}
}
