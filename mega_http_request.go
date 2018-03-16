package main

import (
	"fmt"
	"io/ioutil"
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
	SetHeaders(endpoint.Headers, req)

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

func GetMetricData() (data []byte) {
	var tr *http.Transport

	_u := fmt.Sprintf(APIFormat, Config.KubeSettings.MetricAPIHost, Config.KubeSettings.MetricAPIPort, Config.KubeSettings.MetricAPIPath)

	tr = &http.Transport{}

	tr.ResponseHeaderTimeout = 60 * time.Second
	req, _ := http.NewRequest("GET", _u, nil)

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)

	SetHeaders(APIAuthHeaders, req)
	if err != nil {
		Config.KubeSettings.BadConfig = true
	} else {
		Config.KubeSettings.BadConfig = false
	}

	if resp != nil {
		if resp.Body != nil {
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			data = bodyBytes
			resp.Body.Close()
		}
	}
	tr.CloseIdleConnections()
	return
}

func SetHeaders(headers string, req *http.Request) {
	sets := strings.Split(headers, NewLine)

	for i := range sets {
		split := strings.SplitN(sets[i], HeaderSeparator, 2)
		if len(split) == 2 {
			req.Header.Set(split[0], split[1])
		}
	}
}
