package main

import "time"

var megaWorkspace string = "megaWorkSpace"

var ContentJson = "application/json"
var TwFormat = "%s%s"
var DayInSeconds int64 = 86400

var configName string = "settings.json"
var logDirectory string = "logDirectory"

var logformat string = "log.%s.json"
var urlformat string = "%s%s"

var Production bool = false

const Checkinterval time.Duration = 2 * time.Minute // minutes

//SMTP notifier constants
const DownMsg string = `Dear %s
Server %s(%s) has an uptime below %.2f%%.

Megalith`
const DownSub string = "%s is server down."
const NewLine string = "\n"
const HeaderSeparator string = ":"
const SMTPNoSettingsError string = "SMTP settings are not set."
const EmptyString string = ""
const SMTPMessage string = "From: %s\nTo: %s\nSubject:%s\nMIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n%s"
const SMTPAddress string = "%s:%s"

// HTTP request constants
const MaxPossibleHTTPSuccessCode int = 300
const Zero int = 0
const NetworkAccessErrorCode int = 900

// Timer constants
const BeatAt string = "Beat at "
