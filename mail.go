// Copyright (c) 2023 Julian Müller (ChaoticByte)

package main

import (
	"crypto/tls"
	"fmt"
	"mime"
	"mime/quotedprintable"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
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

func sendNotices(recipient string, notices []*WidNotice, template MailTemplate, auth smtp.Auth, smtpConfig SmtpSettings, cache *map[string][]byte) error {
	logger.debug("Generating and sending mails for recipient " + recipient + " ...")
	cacheHits := 0
	cacheMisses := 0
	mails := [][]byte{}
	for _, n := range notices {
		var data []byte
		cacheResult := (*cache)[n.Uuid]
		if len(cacheResult) > 0 {
			cacheHits++
			data = cacheResult
		} else {
			cacheMisses++
			mailContent, err := template.generate(TemplateData{n, Version})
			if err != nil {
				logger.error("Could not create mail from template")
				logger.error(err)
			}
			// serialize mail
			data = mailContent.serializeValidMail(smtpConfig.From, recipient)
			// add to cache
			(*cache)[n.Uuid] = data
		}
		mails = append(mails, data)
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

func sendMails(smtpConf SmtpSettings, auth smtp.Auth, to string, data [][]byte) error {
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
		logger.debug("Mail Server supports TLS")
	} else {
		logger.debug("Mail Server doesn't support TLS")
	}
	logger.debug("Authenticating to mail server ...")
	err = connection.Auth(auth)
	if err != nil { return err }
	if logger.LogLevel >= 3 {
		fmt.Printf("DEBUG %v Sending mails to server ", time.Now().Format("2006/01/02 15:04:05.000000"))
	}
	for _, d := range data {
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

func mailAddressIsValid(address string) bool {
	_, err := mail.ParseAddress(address);
	return err == nil
}
