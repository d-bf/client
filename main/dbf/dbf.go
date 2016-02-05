package main

import (
	"github.com/d-bf/client/config"
	"github.com/d-bf/client/dbf"
	"github.com/d-bf/client/server"
	"github.com/d-bf/client/term"
	"os"
)

func deferPanic() {
	if panicVal := recover(); panicVal != nil { // Recovering from panic
		if exitCode, ok := panicVal.(int); ok { // Panic value is integer
			os.Exit(exitCode) // Exit with the integer panic value
		}
	}
}

func initialize() {
	defer deferPanic()

	dbf.Init()
	config.Init()
	config.Check()
	server.Init()
}

func main() {
	defer deferPanic()

	initialize()

	term.Clear()

	server.GetVendor("cracker", "hashcat", "cpu_linux_64")
}
