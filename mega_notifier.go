package main

import (
		"net/smtp"
		"fmt"
		"github.com/saintpete/twilio-go"
		"errors"
		"log"
)


// Send email to contact
func SendEmail(subject, body, to string, mail MailSettings) error {

	if mail.Host == "" || mail.Email == "" || mail.Password == "" || mail.Port == "" {
		return errors.New("SMTP settings are not set.")
	}

	msg := fmt.Sprintf("From: %s\nTo: %s\nSubject:%s\nMIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n%s", mail.Email, to, subject, body)

	err := smtp.SendMail(fmt.Sprintf("%s:%s", mail.Host, mail.Port),
		smtp.PlainAuth("", mail.Email, mail.Password, mail.Host),
		mail.Email, []string{to}, []byte(msg))

	if err != nil {
		return err
	}

	return nil
}


// Send sms to contact
func (user Contact) SendSMS(message string) error {
	client := twilio.NewClient(Config.SMS.SID, Config.SMS.Token, nil)
	_, err := client.Messages.SendMessage(Config.SMS.From, fmt.Sprintf(TwFormat, Config.SMS.CountryCode, user.Phone), message, nil)
	return err
}


//Notification dispatcher
func Notify(server Server, contacts []Contact, mailcfg MailSettings) {
	if contacts != nil {
		for _, contact := range contacts {
			if inArr(contact.Watching, server.ID) && contact.Threshold > (server.Uptime*100) {
				if contact.Email != "" {
					err := SendEmail(fmt.Sprintf(DownSub, server.Host), fmt.Sprintf(DownMsg, contact.Nickname, server.Nickname, server.Host, contact.Threshold), contact.Email, mailcfg)
					if err != nil {
						log.Println(err)
					}
				}
				if contact.Phone != "" {
					err := contact.SendSMS(fmt.Sprintf(DownMsg, contact.Nickname, server.Nickname, server.Host, contact.Threshold))
					if err != nil {
						log.Println(err)
					}
				}
			}
		}
	}
}