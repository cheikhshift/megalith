package main

import (
	"errors"
	"fmt"
	"github.com/kevinburke/twilio-go"
	"log"
	"net/smtp"
)

//Notification dispatcher
func Notify(server Server, contacts []Contact, mailcfg MailSettings, smsinfo TwilioInfo) {

	if contacts != nil {
		for _, contact := range contacts {
			if inArr(contact.Watching, server.ID) && contact.Threshold > (server.Uptime*100) {
				var err error
				messageId := fmt.Sprintf(TwFormat, server.ID, DownMsg)
				safetoAlert := ShouldAlert(messageId)

				if !safetoAlert {
					return
				}

				if contact.Email != EmptyString {
					err = SendEmail(fmt.Sprintf(DownSub, server.Host), fmt.Sprintf(DownMsg, contact.Nickname, server.Nickname, server.Host, contact.Threshold), contact.Email, mailcfg)
					if err != nil {
						log.Println(err)
					}
				}
				if contact.Phone != EmptyString {
					err = contact.SendSMS(fmt.Sprintf(DownMsg, contact.Nickname, server.Nickname, server.Host, contact.Threshold), smsinfo)
					if err != nil {
						log.Println(err)
					}
				}

				if err != nil {
					RemoveWithID(server.ID)
				}
			}
		}
	}
}

// Func used to notify users on kubernetes' stats
func NotifyPodContacts(pod PodConfig, contacts []Contact, mailcfg MailSettings, smsinfo TwilioInfo, message string) {
	messageId := fmt.Sprintf(TwFormat, pod.Name, message)

	safetoAlert := ShouldAlert(messageId)
	if !safetoAlert {
		return
	}

	if contacts != nil {
		for _, contact := range contacts {
			if inArr(contact.Pods, pod.Name) {
				var err error
				if contact.Email != EmptyString {
					err := SendEmail(fmt.Sprintf(DownSubk8s, pod.Name), fmt.Sprintf(DownMsgk8s, contact.Nickname, pod.Name, message), contact.Email, mailcfg)
					if err != nil {
						log.Println(err)
					}
				}
				if contact.Phone != EmptyString {
					err := contact.SendSMS(fmt.Sprintf(DownMsgk8s, contact.Nickname, pod.Name, message), smsinfo)
					if err != nil {
						log.Println(err)
					}
				}

				if err != nil {
					RemoveWithID(pod.Name)
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
func (user Contact) SendSMS(message string, smsinfo TwilioInfo) error {
	client := twilio.NewClient(smsinfo.SID, smsinfo.Token, nil)
	_, err := client.Messages.SendMessage(smsinfo.From, fmt.Sprintf(TwFormat, smsinfo.CountryCode, user.Phone), message, nil)
	return err
}
