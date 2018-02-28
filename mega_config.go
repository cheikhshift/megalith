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

const Checkinterval time.Duration = 2 * time.Minute // minutes

const DownMsg string = `Dear %s
Server %s(%s) has an uptime below %.2f%%.

Megalith`
const DownSub string = "%s is server down."



