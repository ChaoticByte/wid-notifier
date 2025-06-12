// Copyright (c) 2023 Julian Müller (ChaoticByte)

package main

import (
	"fmt"
	"net/smtp"
	"os"
	"slices"
	"time"
)

var executableName string
var logger Logger


func showVersion() {
	fmt.Printf("wid-notifier %s\n", Version)
}

func showHelp() {
	fmt.Printf("Usage: %v <configfile>\n\nIf the config file doesn't exist, an incomplete \n" +
			   "configuration with default values is created.\n\n",
			   executableName)
	showVersion()
}

func main() {
	// get cli arguments
	args := os.Args
	executableName = args[0]
	if len(args) < 2 {
		showHelp()
		os.Exit(1)
	}
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			showHelp()
			os.Exit(0)
		} else if arg == "--version" {
			showVersion()
			os.Exit(0)
		}
	}
	configFilePath := args[1]
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
		lastPublished := map[string]time.Time{} // endpoint id : last published timestamp
		cache := map[string][]byte{}            // cache generated emails for reuse
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
				lastPublished[a.Id] = t
			}
		}
		logger.debug(fmt.Sprintf("Got %v new notices", len(newNotices)))
		if len(newNotices) > 0 {
			logger.info("Sending email notifications ...")
			// mail recipient : pointer to slice of wid notices to be sent
			noticesToBeSent := map[string][]*WidNotice{}
			recipientsNotified := 0
			var err error
			for _, l := range *config.Lists {
				// Filter notices for this list
				for _, f := range l.Filter {
					for _, n := range f.filter(newNotices) {
						np := &n
						for _, r := range l.Recipients {
							if !noticeSliceContains(noticesToBeSent[r], np) {
								noticesToBeSent[r] = append(noticesToBeSent[r], np)
							}
						}
					}
				}
			}
			for r, notices := range noticesToBeSent {
				// sort by publish date
				slices.SortFunc(notices, func(a *WidNotice, b *WidNotice) int {
					if a.Published == b.Published {
						return 0
					} else if a.Published.After(b.Published) {
						return 1
					} else {
						return -1
					}
				})
				// send
				err = sendNotices(r, notices, mailTemplate, mailAuth, config.SmtpConfiguration, &cache)
				if err != nil {
					logger.error(err)
				} else {
					recipientsNotified++
				}
			}
			if recipientsNotified < 1 && err != nil {
				logger.error("Couldn't send any mail notification!")
			} else {
				for id, t := range lastPublished {
					persistent.data.(PersistentData).LastPublished[id] = t
					persistent.save()
				}
				logger.info(fmt.Sprintf("Email notifications sent to %v of %v recipients", recipientsNotified, len(noticesToBeSent)))
			}
		}
		dt := int(time.Now().UnixMilli() - t1)
		time.Sleep(time.Millisecond * time.Duration((config.ApiFetchInterval * 1000) - dt))
	}
}
