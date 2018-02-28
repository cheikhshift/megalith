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

	SaveLog(server.ID, &logcurrent)
	for endIndex, endpointCheck := range server.Endpoints {
		apiid := fmt.Sprintf(urlformat, endpointCheck.Method, endpointCheck.Path)
		success, failed := CountAndReturn(logcurrent.Requests, apiid)
		GL.Lock.Lock()
		Config.Servers[servIndex].Endpoints[endIndex].Uptime = float64(success) / float64(success+failed)
		GL.Lock.Unlock()
	}

	success, failed := CountAndReturn(logcurrent.Requests, EmptyString)
	GL.Lock.Lock()
	Config.Servers[servIndex].Uptime = float64(success) / float64(success+failed)
	GL.Lock.Unlock()
	go Notify(Config.Servers[servIndex], Config.Contacts, Config.Mail)
	go SaveConfig(&Config)

}

func CountAndReturn(requests []Request, owner string) (int, int) {
	success := Zero
	failed := Zero

	for _, reqcap := range requests {
		if reqcap.Owner == owner || owner == "" {
			if reqcap.Code < MaxPossibleHTTPSuccessCode {
				success++
			} else {
				failed++
			}
		}
	}
	return success, failed
}
