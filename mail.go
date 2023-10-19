// Copyright (c) 2023 Julian Müller (ChaoticByte)

package main

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/mail"
	"net/smtp"
	"slices"
	"time"
)

const MAIL_LINE_SEP = "\r\n"

type MailContent struct {
	Subject string
	Body string
}

func (c MailContent) serializeValidMail(from string, to string) []byte {
	// We'll send base64 encoded Subject & Body, because we Dschörmäns have umlauts
	// and I'm too lazy to encode ä into =E4 and so on
	subjectEncoded := base64.StdEncoding.EncodeToString([]byte(c.Subject))
	bodyEncoded := base64.StdEncoding.EncodeToString([]byte(c.Body))
	data := []byte(fmt.Sprintf(
		"Content-Type: text/plain; charset=\"utf-8\"\r\nContent-Transfer-Encoding: base64\r\nFrom: %v%vTo: %v%vSubject: =?utf-8?b?%v?=%v%v%v",
		from, MAIL_LINE_SEP,
		to, MAIL_LINE_SEP,
		subjectEncoded, MAIL_LINE_SEP,
		MAIL_LINE_SEP,
		bodyEncoded))
	// done, I guess
	return data
}

type Recipient struct {
	Address string `json:"address"`
	// Must be a configured filter id
	Filters []Filter `json:"include"`
}

type SmtpSettings struct {
	From string `json:"from"`
	ServerHost string `json:"host"`
	ServerPort int `json:"port"`
	User string `json:"user"`
	Password string `json:"password"`
}

func (r Recipient) filterAndSendNotices(notices []WidNotice, template MailTemplate, auth smtp.Auth, smtpConfig SmtpSettings, cache *map[string][]byte) error {
	filteredNotices := []WidNotice{}
	for _, f := range r.Filters {
		for _, n := range f.filter(notices) {
			if !noticeSliceContains(filteredNotices, n) {
				filteredNotices = append(filteredNotices, n)
			}
		}
	}
	slices.Reverse(filteredNotices)
	logger.debug(fmt.Sprintf("Including %v of %v notices for recipient %v", len(filteredNotices), len(notices), r.Address))
	logger.debug("Generating and sending mails to " + r.Address + " ...")
	cacheHits := 0
	cacheMisses := 0
	mails := [][]byte{}
	for _, n := range filteredNotices {
		var data []byte
		cacheResult := (*cache)[n.Uuid]
		if len(cacheResult) > 0 {
			cacheHits++
			data = cacheResult
		} else {
			cacheMisses++
			mailContent, err := template.generate(n)
			if err != nil {
				logger.error("Could not create mail from template")
				logger.error(err)
			}
			// serialize & send mail
			data = mailContent.serializeValidMail(smtpConfig.From, r.Address)
			// add to cache
			(*cache)[n.Uuid] = data
		}
		mails = append(mails, data)
	}
	logger.debug(fmt.Sprintf("%v mail cache hits, %v misses", cacheHits, cacheMisses))
	err := sendMails(
		smtpConfig,
		auth,
		r.Address,
		mails,
	)
	if err != nil {
		return err
	}
	logger.debug("Successfully sent all mails to " + r.Address)
	return nil
}

func sendMails(smtpConf SmtpSettings, auth smtp.Auth, to string, data [][]byte) error {
	addr := fmt.Sprintf("%v:%v", smtpConf.ServerHost, smtpConf.ServerPort)
	logger.debug("Connecting to mail server at " + addr + " ...")
	connection, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer connection.Close()
	// can leave out connection.Hello
	hasTlsExt, _ := connection.Extension("starttls")
	if hasTlsExt {
		err = connection.StartTLS(&tls.Config{ServerName: smtpConf.ServerHost})
		if err != nil {
			return err
		}
		logger.debug("Mail Server supports TLS")
	} else {
		logger.debug("Mail Server doesn't support TLS")
	}
	logger.debug("Authenticating to mail server ...")
	err = connection.Auth(auth)
	if err != nil {
		return err
	}
	if logger.LogLevel >= 3 {
		fmt.Printf("DEBUG %v Sending mails to server ", time.Now().Format("2006/01/02 15:04:05.000000"))
	}
	for _, d := range data {
		err = connection.Mail(smtpConf.From)
		if err != nil {
			return err
		}
		err = connection.Rcpt(to)
		if err != nil {
			return err
		}
		writer, err := connection.Data()
		if err != nil {
			return err
		}
		_, err = writer.Write(d)
		if err != nil {
			return err
		}
		err = writer.Close()
		if err != nil {
			return err
		}
		if logger.LogLevel >= 3 {
			print(".")
		}
	}
	if logger.LogLevel >= 3 {
		print("\n")
	}
	return connection.Quit()
}

func mailAddressIsValid(address string) bool {
	_, err := mail.ParseAddress(address);
	return err == nil
}
