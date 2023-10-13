// Copyright (c) 2023 Julian Müller (ChaoticByte)

package main

import (
	"fmt"
	"net/smtp"
	"os"
	"time"
)

func main() {
	// get cli arguments
	args := os.Args
	if len(args) < 2 {
		fmt.Printf("Usage: %v <configfile>\nIf the config file doesn't exist, a incomplete configuration with default values is created.\n", args[0])
		os.Exit(1)
	}
	configFilePath := os.Args[1]
	// init
	println("INFO\tInitializing ...")
	config := NewDataStore(
		configFilePath,
		NewConfig(),
		true,
		0600,
	).data.(Config)
	persistent := NewDataStore(
		config.PersistentDataFilePath,
		NewPersistentData(config),
		false,
		0640)
	// exit handler
	defer println("INFO\tExiting ...")
	// check config
	checkConfig(config)
	// create mail template from mail template config
	if config.Template.SubjectTemplate == "" {
		config.Template.SubjectTemplate = DEFAULT_SUBJECT_TEMPLATE
	}
	if config.Template.BodyTemplate == "" {
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
				enabledApiEndpoints = append(enabledApiEndpoints, a)
			}
		}
	}
	// main loop
	for {
		t1 := time.Now().UnixMilli()
		newNotices := []WidNotice{}
		for _, a := range enabledApiEndpoints {
			fmt.Printf("INFO\t%v Querying endpoint '%v' for new notices ...\n", time.Now().Format(time.RFC3339Nano), a.Id)
			n, t, err := a.getNotices(persistent.data.(PersistentData).LastPublished[a.Id])
			if err != nil {
				// retry
				n, t, err = a.getNotices(persistent.data.(PersistentData).LastPublished[a.Id])
			}
			if err != nil {
				// ok then...
				fmt.Print("ERROR\t", err)
			} else {
				newNotices = append(newNotices, n...)
				persistent.data.(PersistentData).LastPublished[a.Id] = t
				persistent.save()
			}
		}
		if len(newNotices) > 0 {
			for _, r := range config.Recipients {
				fmt.Printf("INFO\t%v Sending email notifications ...\n", time.Now().Format(time.RFC3339Nano))
				err := r.filterAndSendNotices(newNotices, mailTemplate, mailAuth, config.SmtpConfiguration)
				if err != nil {
					fmt.Printf("ERROR\t%v\n", err)
				} else {
					fmt.Printf("INFO\t%v Email notifications sent.\n", time.Now().Format(time.RFC3339Nano))
				}
			}
		}
		t2 := time.Now().UnixMilli()
		dt := int(t2 - t1)
		time.Sleep(time.Millisecond * time.Duration((config.ApiFetchInterval * 1000) - dt))
	}
}
