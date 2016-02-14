package main

import (
	"fmt"
	"github.com/d-bf/client/dbf"
	"github.com/d-bf/client/term"
	"os"
	"path/filepath"
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

	// Set path data
	pathData, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err == nil {
		dbf.PathData = pathData + string(os.PathSeparator) + "dbf-data" + string(os.PathSeparator)
	} else {
		fmt.Fprintf(os.Stderr, "Can not set current path. Error: %s\n", err)
		panic(1)
	}

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
