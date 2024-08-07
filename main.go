// Copyright (c) 2023 Julian Müller (ChaoticByte)

package main

import (
	"fmt"
	"net/smtp"
	"os"
	"slices"
	"time"
)

var logger Logger

func main() {
	// get cli arguments
	args := os.Args
	if len(args) < 2 {
		fmt.Printf( "Usage: %v <configfile>\n\nIf the config file doesn't exist, an incomplete \n" +
					"configuration with default values is created.\n\nVersion: %s\n",
					args[0],
					Version)
		os.Exit(1)
	}
	configFilePath := os.Args[1]
	// create logger
	logger = NewLogger(2)
	// init
	logger.info("Initializing ...")
	defer logger.info("Exiting ...")
	// open & check config
	config := NewDataStore(
		configFilePath,
		NewConfig(),
		true,
		0600,
	).data.(Config)
	logger.LogLevel = config.LogLevel
	logger.debug("Checking configuration file ...")
	checkConfig(config)
	// create mail template from mail template config
	logger.debug("Parsing mail template ...")
	if config.Template.SubjectTemplate == "" {
		logger.debug("Using default template for mail subject")
		config.Template.SubjectTemplate = DEFAULT_SUBJECT_TEMPLATE
	}
	if config.Template.BodyTemplate == "" {
		logger.debug("Using default template for mail body")
		config.Template.BodyTemplate = DEFAULT_BODY_TEMPLATE
	}
	mailTemplate := NewTemplateFromTemplateConfig(config.Template)
	// mail authentication from config
	mailAuth := smtp.PlainAuth(
		"",
		config.SmtpConfiguration.User,
		config.SmtpConfiguration.Password,
		config.SmtpConfiguration.ServerHost,
	)
	// filter out disabled api endpoints
	enabledApiEndpoints := []ApiEndpoint{}
	for _, a := range apiEndpoints {
		for _, b := range config.EnabledApiEndpoints {
			if a.Id == b {
				logger.debug("Endpoint '" + b + "' is enabled")
				enabledApiEndpoints = append(enabledApiEndpoints, a)
			}
		}
	}
	// open data file
	persistent := NewDataStore(
		config.PersistentDataFilePath,
		NewPersistentData(config),
		false,
		0640)
	// main loop
	logger.debug("Entering main loop ...")
	for {
		t1 := time.Now().UnixMilli()
		newNotices := []WidNotice{}
		cache := map[string][]byte{}
		for _, a := range enabledApiEndpoints {
			logger.info("Querying endpoint '" + a.Id + "' for new notices ...")
			n, t, err := a.getNotices(persistent.data.(PersistentData).LastPublished[a.Id])
			if err != nil {
				// retry (once)
				logger.warn("Couldn't query notices from API endpoint '" + a.Id + "'. Retrying ...")
				logger.warn(err)
				n, t, err = a.getNotices(persistent.data.(PersistentData).LastPublished[a.Id])
			}
			if err != nil {
				// ok then...
				logger.error("Couldn't query notices from API endpoint '" + a.Id + "'")
				logger.error(err)
			} else if len(n) > 0 {
				newNotices = append(newNotices, n...)
				persistent.data.(PersistentData).LastPublished[a.Id] = t
				persistent.save()
			}
		}
		logger.debug(fmt.Sprintf("Got %v new notices", len(newNotices)))
		if len(newNotices) > 0 {
			logger.info("Sending email notifications ...")
			recipientsNotified := 0
			for _, r := range config.Recipients {
				// Filter notices for this recipient
				filteredNotices := []WidNotice{}
				for _, f := range r.Filters {
					for _, n := range f.filter(newNotices) {
						if !noticeSliceContains(filteredNotices, n) {
							filteredNotices = append(filteredNotices, n)
						}
					}
				}
				slices.Reverse(filteredNotices)
				logger.debug(fmt.Sprintf("Including %v of %v notices for recipient %v", len(filteredNotices), len(newNotices), r.Address))
				// Send notices
				err := r.sendNotices(filteredNotices, mailTemplate, mailAuth, config.SmtpConfiguration, &cache)
				if err != nil {
					logger.error(err)
				} else {
					recipientsNotified++
				}
			}
			logger.info(fmt.Sprintf("Email notifications sent to %v of %v recipients", recipientsNotified, len(config.Recipients)))
		}
		dt := int(time.Now().UnixMilli() - t1)
		time.Sleep(time.Millisecond * time.Duration((config.ApiFetchInterval * 1000) - dt))
	}
}
