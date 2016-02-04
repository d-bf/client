package main

import (
	"github.com/d-bf/client/config"
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

func init() {
	defer deferPanic()

	config.Check()
}

func main() {
	defer deferPanic()

	term.Clear()
}