package main

import (
	"github.com/cheikhshift/gos/core"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

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

func InitConfigLoad() {
	if _, err := os.Stat(megaWorkspace); os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Join(megaWorkspace, logDirectory), 0700)
		if err != nil {
			panic(err)
		}
		Config = &MegaConfig{}
		SaveConfig(&Config)
	} else {
		err = LoadConfig(&Config)
		if err != nil {
			panic(err)
		}

	}
}

func LaunchBrowser() {
	Windows := strings.Contains(runtime.GOOS, "windows")
	if Prod {
		if !Windows {
			if isMac := strings.Contains(runtime.GOOS, "arwin"); isMac {
				core.RunCmd(DarwinOpen)
			} else {
				core.RunCmd(LinuxOpen)
			}
		} else {
			core.RunCmd(NTOpen)
		}
	}
}

func ChdirHome() {
	Windows := strings.Contains(runtime.GOOS, "windows")
	if Windows {
		os.Chdir(os.ExpandEnv("$USERPROFILE"))
	} else {
		os.Chdir(os.ExpandEnv("$HOME"))
	}
}
