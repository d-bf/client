package main

import (
	"fmt"
	"github.com/d-bf/client/dbf"
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
	fmt.Println("Initializing...")
	dbf.InitLog()
	dbf.InitConfig()
}

func main() {
	defer deferPanic()

	initialize()
	term.Clear()

	//	for { // Infinite loop
	dbf.GetTask()
	//	}
}
