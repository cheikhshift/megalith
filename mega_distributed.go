package main

import "github.com/theckman/go-flock"
import "fmt"

var WorkerMode = false

var DispatcherAddressPort string = DefaultAddress

var WorkerAddressPort string = DefaultAddress

var Worker Server
var Dispatcher Server

// Alias of Server.

var fileLock = flock.NewFlock(fmt.Sprintf(urlformat, GenConfigName(), LockExt))

func ShouldLock() {
	if !isInContainer {
		GL.Lock.Lock()
	} else {
		fileLock.Lock()
	}
}

func ShouldUnlock() {
	if !isInContainer {
		GL.Lock.Unlock()
	} else {
		fileLock.Unlock()
	}
}

const (
	ProcessServerPath  string = "/momentum/funcs?name=ProcessServer"
	UpdateServerPath   string = "/momentum/funcs?name=UpdateServer"
	RegisterServerPath string = "/momentum/funcs?name=RegisterServer"
)
