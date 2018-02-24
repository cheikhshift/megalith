package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ContentJson = "application/json"

func SendEmail(subject, body, to string) error {

	if Config.Mail.Host == "" || Config.Mail.Email == "" || Config.Mail.Password == "" || Config.Mail.Port == "" {
		return errors.New("SMTP settings are not set.")
	}

	msg := fmt.Sprintf("From: %s\nTo: %s\nSubject:%s\nMIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n%s", Config.Mail.Email, to, subject, body)

	err := smtp.SendMail(fmt.Sprintf("%s:%s", Config.Mail.Host, Config.Mail.Port),
		smtp.PlainAuth("", Config.Mail.Email, Config.Mail.Password, Config.Mail.Host),
		Config.Mail.Email, []string{to}, []byte(msg))

	if err != nil {
		return err
	}

	return nil
}

func SaveConfig(v interface{}) error {
	str := mResponse(v)
	pathoffile := filepath.Join(megaWorkspace, configName)
	strbytes := []byte(str)
	err := ioutil.WriteFile(pathoffile, strbytes, 0700)
	strbytes = nil
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

	pathoffile := filepath.Join(megaWorkspace, configName)

	data, err := ioutil.ReadFile(pathoffile)
	if err != nil {
		return err
	}
	strdata := string(data)
	bts := []byte(strdata)
	err = json.Unmarshal(bts, targ)
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
	GL.Lock.Lock()
	for servIndex, server := range Config.Servers {
		Process(server, servIndex)
	}
	GL.Lock.Unlock()
}

func Process(server Server, servIndex int) {
	if server.Live {
		logcurrent := RequestLog{}

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
					if reqcap.Code == 200 {
						success++
					} else {
						failed++
					}
				}

			}
			Config.Servers[servIndex].Endpoints[endIndex].Uptime = float64(success) / float64(success+failed)

		}
		success := 0
		failed := 0
		for _, reqcap := range logcurrent.Requests {
			if reqcap.Code == 200 {
				success++
			} else {
				failed++
			}
		}
		Config.Servers[servIndex].Uptime = float64(success) / float64(success+failed)

		Notify(Config.Servers[servIndex])
		SaveConfig(Config)
	}
}

func Notify(server Server) {
	for _, contact := range Config.Contacts {
		if inArr(contact.Watching, server.ID) && contact.Threshold > server.Uptime {
			err := SendEmail(fmt.Sprintf(DownSub, server.Host), fmt.Sprintf(DownMsg, contact.Nickname, server.Nickname, server.Host, contact.Threshold), contact.Email)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func Req(server Server, endpoint Endpoint) int {
	var tr *http.Transport
	_u := fmt.Sprintf(urlformat, server.Host, endpoint.Path)

	tr = &http.Transport{}

	tr.IdleConnTimeout = endpoint.Timeout * time.Second

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
	return ert
}
