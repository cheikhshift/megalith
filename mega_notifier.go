package main

import (
	"errors"
	"fmt"
	"github.com/saintpete/twilio-go"
	"log"
	"net/smtp"
)

//Notification dispatcher
func Notify(server Server, contacts []Contact, mailcfg MailSettings) {
	if contacts != nil {
		for _, contact := range contacts {
			if inArr(contact.Watching, server.ID) && contact.Threshold > (server.Uptime*100) {
				if contact.Email != EmptyString {
					err := SendEmail(fmt.Sprintf(DownSub, server.Host), fmt.Sprintf(DownMsg, contact.Nickname, server.Nickname, server.Host, contact.Threshold), contact.Email, mailcfg)
					if err != nil {
						log.Println(err)
					}
				}
				if contact.Phone != EmptyString {
					err := contact.SendSMS(fmt.Sprintf(DownMsg, contact.Nickname, server.Nickname, server.Host, contact.Threshold))
					if err != nil {
						log.Println(err)
					}
				}
			}
		}
	}
}

// Send email to contact
func SendEmail(subject, body, to string, mail MailSettings) error {

	if mail.Host == "" || mail.Email == "" || mail.Password == "" || mail.Port == "" {
		return errors.New(SMTPNoSettingsError)
	}

	msg := fmt.Sprintf(SMTPMessage, mail.Email, to, subject, body)

	err := smtp.SendMail(fmt.Sprintf(SMTPAddress, mail.Host, mail.Port),
		smtp.PlainAuth(EmptyString, mail.Email, mail.Password, mail.Host),
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
