package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

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
