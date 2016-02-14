package main

import (
	"fmt"
	"github.com/d-bf/client/dbf"
	"github.com/d-bf/client/term"
	"os"
	"time"
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

	for { // Infinite loop
		fmt.Println("Checking for new task from server...")
		dbf.GetTask()
		fmt.Println("Wait before checking for next task...")
		time.Sleep(15 * time.Second)
		term.Clear()
	}
}
