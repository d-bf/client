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

	// Set data path
	pathData, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err == nil {
		dbf.PathData = pathData + string(os.PathSeparator) + "dbf-data" + string(os.PathSeparator)
	} else {
		fmt.Fprintf(os.Stderr, "Can not set current path. Error: %s\n", err)
		panic(1)
	}

	// Check data folder
	if _, err = os.Stat(dbf.PathData); err != nil {
		if os.IsNotExist(err) { // Does not exist, so create it
			if err = os.MkdirAll(dbf.PathData, 0775); err != nil {
				fmt.Fprintf(os.Stderr, "Can not create data folder. Error: %s\n", err) // Error in creating
				panic(1)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Can not access data folder. Error: %s\n", err) // Error in accessing
			panic(1)
		}
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
