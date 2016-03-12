package main

import (
	"fmt"
	"github.com/d-bf/client/dbf"
	"github.com/d-bf/client/term"
	"os"
	"path/filepath"
	"time"
)

var timer uint

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

	term.Clear()
	initialize()
	term.Clear()

	timer = 0
	dbf.ResetTimer = false

	var enterPressed bool = false
	go func(enterPressed *bool) {
		for {
			fmt.Scanln()
			*enterPressed = true
		}
	}(&enterPressed)

	for { // Infinite loop
		fmt.Println("Checking for new task from server...")

		dbf.GetTask()

		fmt.Println("Done\n")

		timeout := getTimer()
		enterPressed = false
		showRemainingTime(timeout)
		timeout--

		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			showRemainingTime(timeout)

			if enterPressed {
				ticker.Stop()
				dbf.ResetTimer = true
				enterPressed = false
				break
			}

			if timeout == 0 {
				ticker.Stop()
				break
			}

			timeout--
		}

		term.Clear()
	}
}

func getTimer() uint {
	if dbf.ResetTimer {
		timer = 1
		dbf.ResetTimer = false
	} else {
		timer++
	}

	if timer > 150 {
		timer = 150
	}

	return timer * 4
}

func showRemainingTime(timeout uint) {
	switch {
	case timeout > 119:
		fmt.Printf("\rPerform next check in %d minutes... (Press enter to check now) ", timeout/60)
	case timeout > 59:
		fmt.Printf("\rPerform next check in %d minute... (Press enter to check now) ", 1)
	case timeout > 1:
		fmt.Printf("\rPerform next check in %d seconds... (Press enter to check now) ", timeout)
	default:
		fmt.Printf("\rPerform next check in %d second... (Press enter to check now) ", 1)
	}
}
