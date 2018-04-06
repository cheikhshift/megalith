package main

import "fmt"

func Process(server Server, servIndex int) {

	logcurrent := RequestLog{}
	ShouldDeleteLog(server.ID)
	LoadLog(server.ID, &logcurrent)

	for _, endpointCheck := range server.Endpoints {
		reqframe := Req(server, endpointCheck)
		apiid := fmt.Sprintf(urlformat, endpointCheck.Method, endpointCheck.Path)
		logcurrent.Requests = append(logcurrent.Requests, Request{Code: reqframe, Owner: apiid})
	}

	// Lock log file

	SaveLog(server.ID, &logcurrent)

	for endIndex, endpointCheck := range server.Endpoints {
		apiid := fmt.Sprintf(urlformat, endpointCheck.Method, endpointCheck.Path)
		success, failed := CountAndReturn(logcurrent.Requests, apiid)
		ShouldLock()
		Config.Servers[servIndex].Endpoints[endIndex].Uptime = float64(success) / float64(success+failed)
		ShouldUnlock()
	}

	success, failed := CountAndReturn(logcurrent.Requests, EmptyString)
	ShouldLock()
	Config.Servers[servIndex].Uptime = float64(success) / float64(success+failed)
	latestServerInformation := Config.Servers[servIndex]
	ShouldUnlock()

	if Config.AlertsHistory == nil {
		ClearHistory()
	}

	go Notify(latestServerInformation, Config.Contacts, Config.Mail, Config.SMS)
	go SaveConfig(&Config)
}

func CountAndReturn(requests []Request, owner string) (success int, failed int) {
	for _, reqcap := range requests {
		if reqcap.Owner == owner || owner == EmptyString {
			if reqcap.Code < MaxPossibleHTTPSuccessCode {
				success++
			} else {
				failed++
			}
		}
	}
	return success, failed
}
