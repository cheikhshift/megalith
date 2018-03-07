package main

import "time"
import "sync"

type PayloadOfRequest struct {
	req string
}

type MegaConfig struct {
	Mail      MailSettings
	Servers   []Server
	Cl        Clock
	Contacts  []Contact
	SMS       TwilioInfo
	Misc      Settings
	LastReset int64
}
type TrLock struct {
	Lock *sync.RWMutex
}
type Server struct {
	Host, Image, Nickname string
	Endpoints             []Endpoint
	Live                  bool
	ID                    string
	Uptime                float64
}
type Endpoint struct {
	Uptime                      float64
	Method, Path, Headers, Data string
	Timeout                     time.Duration
}
type RequestLog struct {
	Requests []Request
}
type Request struct {
	Code  int
	Owner string
}
type Contact struct {
	Nickname, Email string
	Threshold       float64
	Watching        []string
	ID, Phone       string
}

type MailSettings struct {
	Email, Password, Host, Port string
}

type Settings struct {
	ResetInterval int64
}
type Clock struct {
	Interval int
}

type TwilioInfo struct {
	Token, SID, From, CountryCode string
}
