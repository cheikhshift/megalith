package main

var WorkerMode = false

var DispatcherAddressPort string

var WorkerAddressPort string

var Worker Server
var Dispatcher Server

// Alias of Server.

const (
	ProcessServerPath  string = "/momentum/funcs?name=ProcessServer"
	UpdateServerPath   string = "/momentum/funcs?name=UpdateServer"
	RegisterServerPath string = "/momentum/funcs?name=RegisterServer"
)
