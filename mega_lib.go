package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/saintpete/twilio-go"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func MegaTimer(ticker *time.Ticker) {
	for t := range ticker.C {
		log.Println("Beat at", t)
		go Pulse()
	}
}

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

func SaveConfig(v interface{}) error {
	GL.Lock.Lock()
	str := mResponse(v)
	pathoffile := filepath.Join(megaWorkspace, configName)
	strbytes := []byte(str)
	err := ioutil.WriteFile(pathoffile, strbytes, 0700)
	strbytes = nil
	GL.Lock.Unlock()
	return err

}

func inArr(arr []string, lookup string) (res bool) {
	for _, val := range arr {
		if val == lookup {
			res = true
		}
	}
	return
}

func DeleteLog(name string) {
	pathoffile := filepath.Join(megaWorkspace, logDirectory, name)
	os.Remove(pathoffile)
}

func SaveLog(name string, v interface{}) error {
	str := mResponse(v)
	pathoffile := filepath.Join(megaWorkspace, logDirectory, name)
	strbytes := []byte(str)
	err := ioutil.WriteFile(pathoffile, strbytes, 0700)
	strbytes = nil
	return err
}

func LoadConfig(targ interface{}) error {
	GL.Lock.Lock()
	pathoffile := filepath.Join(megaWorkspace, configName)

	data, err := ioutil.ReadFile(pathoffile)
	if err != nil {
		return err
	}
	strdata := string(data)
	bts := []byte(strdata)
	err = json.Unmarshal(bts, targ)
	GL.Lock.Unlock()
	return err
}

func LoadLog(name string, targ interface{}) error {

	pathoffile := filepath.Join(megaWorkspace, logDirectory, name)

	data, err := ioutil.ReadFile(pathoffile)
	if err != nil {
		return err
	}
	strdata := string(data)
	bts := []byte(strdata)
	err = json.Unmarshal(bts, targ)
	return err
}

func Pulse() {
	if Config.Servers != nil {
		for servIndex, server := range Config.Servers {
			if server.Live {
				go Process(server, servIndex)
			}
		}
	}
}

func ShouldDeleteLog(server string) {
	now := time.Now().Unix()
	if Config.Misc.ResetInterval == 0 {
		return
	}
	if (now - Config.LastReset) > (DayInSeconds * Config.Misc.ResetInterval) {
		DeleteLog(server)
	}
}

func Process(server Server, servIndex int) {

	logcurrent := RequestLog{}
	ShouldDeleteLog(server.ID)
	LoadLog(server.ID, &logcurrent)

	for _, endpointCheck := range server.Endpoints {
		reqframe := Req(server, endpointCheck)
		apiid := fmt.Sprintf(urlformat, endpointCheck.Method, endpointCheck.Path)
		logcurrent.Requests = append(logcurrent.Requests, Request{Code: reqframe, Owner: apiid})
	}

	SaveLog(server.ID, &logcurrent)
	for endIndex, endpointCheck := range server.Endpoints {
		success := 0
		failed := 0
		apiid := fmt.Sprintf(urlformat, endpointCheck.Method, endpointCheck.Path)
		for _, reqcap := range logcurrent.Requests {
			if reqcap.Owner == apiid {
				if reqcap.Code < 300 {
					success++
				} else {
					failed++
				}
			}

		}
		GL.Lock.Lock()
		Config.Servers[servIndex].Endpoints[endIndex].Uptime = float64(success) / float64(success+failed)
		GL.Lock.Unlock()
	}
	success := 0
	failed := 0
	for _, reqcap := range logcurrent.Requests {
		if reqcap.Code < 300 {
			success++
		} else {
			failed++
		}
	}
	GL.Lock.Lock()
	Config.Servers[servIndex].Uptime = float64(success) / float64(success+failed)
	GL.Lock.Unlock()
	go Notify(Config.Servers[servIndex], Config.Contacts, Config.Mail)
	go SaveConfig(&Config)

}

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

func (user Contact) SendSMS(message string) error {
	client := twilio.NewClient(Config.SMS.SID, Config.SMS.Token, nil)
	_, err := client.Messages.SendMessage(Config.SMS.From, fmt.Sprintf(TwFormat, Config.SMS.CountryCode, user.Phone), message, nil)
	return err
}

func Req(server Server, endpoint Endpoint) int {
	var tr *http.Transport
	_u := fmt.Sprintf(urlformat, server.Host, endpoint.Path)

	tr = &http.Transport{}

	tr.ResponseHeaderTimeout = endpoint.Timeout * time.Second
	//tr.Dial = endpoint.Timeout * time.Second

	// wrap parameters around bson.M map under
	// key `req`
	requestBodyReader := strings.NewReader(endpoint.Data)
	req, _ := http.NewRequest(endpoint.Method, _u, requestBodyReader)
	sets := strings.Split(endpoint.Headers, "\n")

	//Split incoming header string by \n and build header pairs
	for i := range sets {
		split := strings.SplitN(sets[i], ":", 2)
		if len(split) == 2 {
			req.Header.Set(split[0], split[1])
		}
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	var ert int
	if err != nil {
		ert = 900
		log.Println(err)
	} else {
		ert = resp.StatusCode
	}
	if resp.Body != nil {
		resp.Body.Close()
	}
	tr.CloseIdleConnections()
	return ert
}
