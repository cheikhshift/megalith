package main

import (
	"log"
	"time"
)

func MegaTimer(ticker *time.Ticker) {
	for t := range ticker.C {
		log.Println(BeatAt, t)
		go Pulse()
		go k8sMonitor()
	}
}

func Pulse() {
	if Config.Servers != nil {
		for servIndex, server := range Config.Servers {
			if server.Live {
				if Worker.Host != "" {
					go DispatchtoWorker(server)
				} else {
					go Process(server, servIndex)
				}
			}
		}
	}
}

func k8sMonitor() {
	ShouldLock()
	tempCopy := Config.KubeSettings
	tempContacts := Config.Contacts
	tempTwConfig := Config.SMS
	tempMailConfig := Config.Mail
	ShouldRequest := Config.KubeSettings.MetricAPIHost != "" && Config.KubeSettings.MetricAPIPath != "" && Config.KubeSettings.MetricAPIPort != ""
	ShouldUnlock()
	if ShouldRequest {
		list, err := RequestPods()
		if err != nil {
			log.Println(err)
		} else {
			for _, pod := range tempCopy.Monitoring {
				// Check and alert contacts
				go list.CheckContainers(pod, tempContacts, tempMailConfig, tempTwConfig)
			}
		}
	}

}
