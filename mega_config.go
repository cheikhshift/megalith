package main

import "time"

// Megalith directory names
var (
	// name of megalith workspace directory.
	// relative to your $HOME directory on Unix/Linux or
	// %USERPROFILE% on Windows.
	// Prefix this variable with a root location to save
	// megalith data in another folder.
	megaWorkspace string = "megaWorkSpace"
	// name of megalith log directory.
	logDirectory string = "logDirectory"
)

// Name of megalith settings file.
var configName string = "settings.json"

// Run megalith in production mode.
var Production bool = false

// Interval to perform requests
// to the servers your are
// monitoring.
const Checkinterval time.Duration = 2 * time.Minute // minutes

// String formats
const (
	//log file name format
	logformat string = "log.%s.json"
	// URL format : HOST_NAME + PATH
	urlformat string = "%s%s"
	// Twilio Phone format : AREA_CODE + PHONE_NUMBER
	TwFormat = "%s%s"
)

//Notifier message constants
const (
	DownMsg string = `Dear %s
Server %s(%s) has an uptime below %.2f%%.

Megalith`
	DownSub string = "%s is down."
)

// Notifier SMTP settings
const (
	SMTPNoSettingsError string = "SMTP settings are not set."
	SMTPMessage         string = "From: %s\nTo: %s\nSubject:%s\nMIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n%s"
	SMTPAddress         string = "%s:%s"
)

// HTTP request constants
const (
	//request considered successful if
	// status code is below 300.
	MaxPossibleHTTPSuccessCode int = 300
	// Network error status code.
	NetworkAccessErrorCode int = 900
)

// Timer constants
const BeatAt string = "Beat at "

// Browser open command
const (
	DarwinOpen string = "open http://localhost:9001/index"
	NTOpen     string = "cmd /C start http://localhost:9001/index"
	LinuxOpen  string = "xdg-open http://localhost:9001/index"
)

// Misc constants
const (
	NewLine         string = "\n"
	HeaderSeparator string = ":"
	EmptyString     string = ""
	ContentJson     string = "application/json"
	DayInSeconds    int64  = 86400
	Zero            int    = 0
	PORT            string = "PORT"
	DefaultAddress  string = "http://127.0.0.1:9001"
	LockExt         string = ".lock"
	DefaultPort     string = "9001"
)

// Misc variables
var (
	OK = []byte("OK")
)

// Headers to use during instance communication
// Intended to be used to provide authentication information
// To access megalith worker cluster/server.
// ie : Content-Type: value\nX-Header: value\n
var AuthorizeHeader string
