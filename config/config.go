package config

import (
	"fmt"
	"github.com/d-bf/client/dbf"
	"os"
	"path/filepath"
)

var (
	DbfConfig    *DbfConf
	PathCurrent  string
	PathVendor   string
	PathCrack    string
	pathConfDir  string
	pathConfFile string
)

func setPathCurrent() {

}

func Init() {
	// Set current path
	var err error
	PathCurrent, err = filepath.Abs(filepath.Dir(os.Args[0]))
	if err == nil {
		PathCurrent += string(os.PathSeparator)
	} else {
		dbf.Log.Printf("%s\n", err)
		panic(1)
	}

	pathConfDir = PathCurrent + "config" + string(os.PathSeparator)
	pathConfFile = pathConfDir + "dbf.json"
	PathVendor = PathCurrent + "vendor" + string(os.PathSeparator)
	PathCrack = PathCurrent + "crack" + string(os.PathSeparator)
}

func checkDir(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) { // Does not exist, so create it
			if err = os.MkdirAll(path, 0775); err != nil {
				dbf.Log.Printf("%s\n", err) // Error in creating
				return err
			}
		} else {
			dbf.Log.Printf("%s\n", err) // Error in accessing
			return err
		}
	}

	return nil
}

func Check() {
	err := checkDir(pathConfDir)
	if err != nil {
		dbf.Log.Printf("%s\n", err)
		panic(1)
	}

	// Check config file
	if _, err := os.Stat(pathConfFile); err != nil {
		if os.IsNotExist(err) { // Does not exist, so create it
			// Create initial config file
			err = createDbfConf()
			if err == nil {
				fmt.Printf("Please enter server's URL in url_api in config file: %s\n", pathConfFile)
				panic(0)
			} else {
				dbf.Log.Printf("%s\n", err)
				panic(1)
			}
		} else {
			dbf.Log.Printf("%s\n", err) // Error in accessing
			panic(1)
		}
	} else { // Sync config file
		err := checkDir(PathVendor)
		if err != nil {
			dbf.Log.Printf("%s\n", err)
			panic(1)
		}

		err = checkDir(PathCrack)
		if err != nil {
			dbf.Log.Printf("%s\n", err)
			panic(1)
		}

		DbfConfig = readDbfConf()

		// Check default vendor files
		for _, platform := range *DbfConfig.Platform {
			_ = platform.Id
		}
	}
}
