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

	requestBodyReader := strings.NewReader(endpoint.Data)
	req, _ := http.NewRequest(endpoint.Method, _u, requestBodyReader)
	sets := strings.Split(endpoint.Headers, NewLine)

	for i := range sets {
		split := strings.SplitN(sets[i], HeaderSeparator, 2)
		if len(split) == 2 {
			req.Header.Set(split[0], split[1])
		}
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	var ert int
	if err != nil {
		ert = NetworkAccessErrorCode
		log.Println(err)
	} else {
		ert = resp.StatusCode
	}
	if resp != nil {
		if resp.Body != nil {
			resp.Body.Close()
		}
	}
	tr.CloseIdleConnections()
	return ert
}
