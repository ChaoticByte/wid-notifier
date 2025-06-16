// +build debug_mail_transfer
// Copyright (c) 2023 Julian Müller (ChaoticByte)

package main

import (
	"fmt"
	"net/smtp"
)

func sendMails(smtpConf SmtpSettings, auth smtp.Auth, to string, mails []*MailContent) error {
	logger.warn("Mail Transfer Debugging is active. Not connecting.")
	logger.info("MAIL TRANSFER: \n\n")
	for _, mc := range mails {
		// serialize mail
		d := mc.serializeValidMail(smtpConf.From, to)
		// output mail
		fmt.Println("MAIL FROM:" + smtpConf.From)
		fmt.Println("RCPT TO:" + to)
		fmt.Println("DATA")
		fmt.Println(string(d))
		fmt.Println(".")
	}
	fmt.Print("\n\n")
	return nil
}
