package main

import (
	"log"
	"time"
)

func MegaTimer(ticker *time.Ticker) {
	for t := range ticker.C {
		log.Println(BeatAt, t)
		go Pulse()
	}
}

func Pulse() {
	if Config.Servers != nil {
		for servIndex, server := range Config.Servers {
			if server.Live {
				go Process(server, servIndex)
			}
		}
	}
}
