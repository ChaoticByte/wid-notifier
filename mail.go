// Copyright (c) 2023 Julian Müller (ChaoticByte)

package main

import (
	"encoding/base64"
	"fmt"
	"net/mail"
	"net/smtp"
	"slices"
	"strings"
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
	bodyEncoded := base64.StdEncoding.EncodeToString(
		[]byte( // ensure that all lines end with CRLF
			strings.ReplaceAll(
				strings.ReplaceAll(c.Body, "\n", MAIL_LINE_SEP), "\r\r", "\r",
			),
		),
	)
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

func (r Recipient) filterAndSendNotices(notices []WidNotice, template MailTemplate, auth smtp.Auth, smtpConfig SmtpSettings) error {
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
	for _, n := range filteredNotices {
		mailContent, err := template.generate(n)
		if err != nil {
			logger.error("Could not create mail from template")
			logger.error(err)
		}
		// serialize & send mail
		data := mailContent.serializeValidMail(smtpConfig.From, r.Address)
		err = smtp.SendMail(
			fmt.Sprintf("%v:%v", smtpConfig.ServerHost, smtpConfig.ServerPort),
			auth,
			smtpConfig.From,
			[]string{r.Address},
			data,
		)
		if err != nil {
			return err
		}
	}
	logger.debug("Successfully sent all mails to " + r.Address)
	return nil
}

func mailAddressIsValid(address string) bool {
	_, err := mail.ParseAddress(address);
	return err == nil
}
