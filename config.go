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
	Recipients []Recipient `json:"recipients"`
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
		Recipients: []Recipient{},
		SmtpConfiguration: SmtpSettings{
			From: "from@example.org",
			User: "from@example.org",
			Password: "SiEhAbEnMiChInSgEsIcHtGeFiLmTdAsDüRfEnSiEnIcHt",
			ServerHost: "example.org",
			ServerPort: 587},
		Template: MailTemplateConfig{
			SubjectTemplate: "",
			BodyTemplate: "",
		},
	}
	return c
}

func checkConfig(config Config) {
	if len(config.Recipients) < 1 {
		logger.error("Configuration is incomplete")
		panic(errors.New("no recipients are configured"))
	}
	for _, r := range config.Recipients {
		if !mailAddressIsValid(r.Address) {
			logger.error("Configuration includes invalid data")
			panic(errors.New("'" + r.Address + "' is not a valid e-mail address"))
		}
		if len(r.Filters) < 1 {
			logger.error("Configuration is incomplete")
			panic(errors.New("recipient " + r.Address + " has no filter defined - at least {'any': true/false} should be configured"))
		}
	}
	if !mailAddressIsValid(config.SmtpConfiguration.From) {
		logger.error("Configuration includes invalid data")
		panic(errors.New("'" + config.SmtpConfiguration.From + "' is not a valid e-mail address"))
	}
}
