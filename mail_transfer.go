// +build !debug_mail_transfer
// Copyright (c) 2023 Julian Müller (ChaoticByte)

package main

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"time"
)


func sendMails(smtpConf SmtpSettings, auth smtp.Auth, to string, mails []*MailContent) error {
	addr := fmt.Sprintf("%v:%v", smtpConf.ServerHost, smtpConf.ServerPort)
	logger.debug("Connecting to mail server at " + addr + " ...")
	connection, err := smtp.Dial(addr)
	if err != nil { return err }
	defer connection.Close()
	// can leave out connection.Hello
	hasTlsExt, _ := connection.Extension("starttls")
	if hasTlsExt {
		err = connection.StartTLS(&tls.Config{ServerName: smtpConf.ServerHost})
		if err != nil { return err }
		logger.debug("Mail Server supports StartTLS")
	} else {
		logger.debug("Mail Server doesn't support StartTLS")
	}
	logger.debug("Authenticating to mail server ...")
	err = connection.Auth(auth)
	if err != nil { return err }
	if logger.LogLevel >= 3 {
		fmt.Printf("DEBUG %v Sending mails to server ", time.Now().Format("2006/01/02 15:04:05.000000"))
	}
	for _, mc := range mails {
		// serialize mail
		d := mc.serializeValidMail(smtpConf.From, to)
		// send mail
		err = connection.Mail(smtpConf.From)
		if err != nil { return err }
		err = connection.Rcpt(to)
		if err != nil { return err }
		writer, err := connection.Data()
		if err != nil { return err }
		_, err = writer.Write(d)
		if err != nil { return err }
		err = writer.Close()
		if err != nil { return err }
		if logger.LogLevel >= 3 {
			print(".")
		}
	}
	if logger.LogLevel >= 3 {
		print("\n")
	}
	return connection.Quit()
}
