package config

import (
	"github.com/d-bf/client/dbf"
	"os"
	"path/filepath"
)

var (
	CurrentPath string
	confPath    string
)

func init() {
	// Set current path
	var err error
	CurrentPath, err = filepath.Abs(filepath.Dir(os.Args[0]))
	if err == nil {
		CurrentPath += string(os.PathSeparator)
	} else {
		dbf.Log.Printf("%s\n", err)
		os.Exit(1)
	}
}

func Check() {
	// Check config dir
	confPath = CurrentPath + "conf" + string(os.PathSeparator)
	if _, err := os.Stat(confPath); err != nil {
		if os.IsNotExist(err) { // Does not exist, so create it
			if err = os.MkdirAll(confPath, 0775); err != nil {
				dbf.Log.Printf("Error 2.2: %s\n", err) // Error in creating
				os.Exit(1)
			}
		} else {
			dbf.Log.Printf("%s\n", err) // Error in accessing
			os.Exit(1)
		}
	}

	// Check config file
	confPath = confPath + "dbf.json"
	if _, err := os.Stat(confPath); err != nil {
		if os.IsNotExist(err) { // Does not exist, so create it
			// Create initial config file
			createDbfConf()
		} else {
			dbf.Log.Printf("%s\n", err) // Error in accessing
			os.Exit(1)
		}
	} else {
		// Sync dbf config
	}
}
