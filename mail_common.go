// Copyright (c) 2023 Julian Müller (ChaoticByte)

package main

import (
	"fmt"
	"mime"
	"mime/quotedprintable"
	"net/mail"
	"net/smtp"
	"strings"
)

type MailContent struct {
	Subject string
	Body string
}

func (c MailContent) serializeValidMail(from string, to string) []byte {
	// format subject using Q Encoding from RFC2047
	subjectEncoded := mime.QEncoding.Encode("utf-8", c.Subject)
	// format body using Quoted-Printable Encoding from RFC2045
	var bodyEncoded strings.Builder
	bew := quotedprintable.NewWriter(&bodyEncoded)
	bew.Write([]byte(c.Body))
	bew.Close()
	// glue it all together
	data := fmt.Appendf(nil, 
		"Content-Type: text/plain; charset=\"utf-8\"\r\nContent-Transfer-Encoding: Quoted-Printable\r\nFrom: %v\r\nTo: %v\r\nSubject: %v\r\n\r\n%v",
		from, to, subjectEncoded, bodyEncoded.String(),
	)
	return data
}

type NotifyList struct {
	Name string `json:"name"`
	Recipients []string `json:"recipients"`
	// Must be a configured filter id
	Filter []Filter `json:"filter"`
}

type SmtpSettings struct {
	From string `json:"from"`
	ServerHost string `json:"host"`
	ServerPort int `json:"port"`
	User string `json:"user"`
	Password string `json:"password"`
}

func sendNotices(recipient string, notices []*WidNotice, template MailTemplate, auth smtp.Auth, smtpConfig SmtpSettings, mailContentCache *map[string]*MailContent) error {
	logger.debug("Generating and sending mails for recipient " + recipient + " ...")
	cacheHits := 0
	cacheMisses := 0
	mails := []*MailContent{}
	for _, n := range notices {
		var mc *MailContent
		cacheResult := (*mailContentCache)[n.Uuid]
		if cacheResult != nil {
			cacheHits++
			mc = cacheResult
		} else {
			cacheMisses++
			mc_, err := template.generate(TemplateData{n, Version})
			if err != nil {
				logger.error("Could not create mail from template")
				logger.error(err)
			} else {
				mc = &mc_
				// add to cache
				(*mailContentCache)[n.Uuid] = mc
			}
		}
		mails = append(mails, mc)
	}
	logger.debug(fmt.Sprintf("%v mail cache hits, %v misses", cacheHits, cacheMisses))
	err := sendMails(
		smtpConfig,
		auth,
		recipient,
		mails,
	)
	if err != nil { return err }
	logger.debug("Successfully sent all mails to " + recipient)
	return nil
}

func mailAddressIsValid(address string) bool {
	_, err := mail.ParseAddress(address);
	return err == nil
}
