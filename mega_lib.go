package main

import "time"

func inArr(arr []string, lookup string) (res bool) {
	for _, val := range arr {
		if val == lookup {
			res = true
		}
	}
	return
}


//Verify if the last time your 
// logs were reset
// exceeds the set period and delete the server's.
// request log.
func ShouldDeleteLog(server string) {
	now := time.Now().Unix()
	if Config.Misc.ResetInterval == 0 {
		return
	}
	if (now - Config.LastReset) > (DayInSeconds * Config.Misc.ResetInterval) {
		DeleteLog(server)
	}
}

