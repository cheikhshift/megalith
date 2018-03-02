package main

import "log"

func FindServer(req string) (server Server, index int) {
	for index, server = range Config.Servers {
		if server.ID == req {
			break
		}
	}
	return
}

func RespondToDispatcher(server Server) {
	request := make(map[string]interface{})
	request["req"] = server
	serverJson := mResponse(request)
	endpoint := Endpoint{Method: "POST", Data: serverJson, Headers: AuthorizeHeader, Path: UpdateServerPath}
	responseCode := Req(Dispatcher, endpoint)
	if responseCode != 200 {
		log.Println("Could not respond to dispatcher.")
	}
}

func DispatchtoWorker(server Server) {
	request := make(map[string]interface{})
	request["req"] = server.ID
	serverJson := mResponse(request)
	endpoint := Endpoint{Method: "POST", Data: serverJson, Headers: AuthorizeHeader, Path: ProcessServerPath}
	responseCode := Req(Worker, endpoint)
	if responseCode != 200 {
		log.Println("Could not dispatch task to worker. Reverting to this instance.")
		server, index := FindServer(server.ID)
		Process(server, index)
	}
}

func SelfAnnounce(addr string) {
	request := make(map[string]interface{})
	request["req"] = WorkerAddressPort

	// if left empty default values
	// from mega_config will be used.
	if addr != "" {
		Dispatcher.Host = addr
	}
	serverJson := mResponse(request)
	endpoint := Endpoint{Method: "POST", Data: serverJson, Headers: AuthorizeHeader, Path: RegisterServerPath}
	responseCode := Req(Dispatcher, endpoint)
	if responseCode != 200 {
		log.Println("Could not reach your dispatcher.")
	}
}

func RegisterWorker(addr string) {
	GL.Lock.Lock()
	Worker = Server{Host: addr}
	log.Println("Worker location address set to :", addr)
	GL.Lock.Unlock()
}
