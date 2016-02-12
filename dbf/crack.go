package dbf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	_CRACK_TYPE_EMBED  = 1
	_CRACK_TYPE_STDIN  = 2
	_CRACK_TYPE_INFILE = 3
)

type StructCrack struct {
	Id            string `json:"id"`
	Generator     string `json:"generator"`
	Cracker       string `json:"cracker"`
	Cmd_generator string `json:"cmd_generator"`
	Cmd_cracker   string `json:"cmd_cracker"`
	Target        string `json:"target"`
}

func processCrack(task *StructCrackTask, crackInfoPath *string) bool {
	var generatorPath, crackerPath, cmdCracker, cmdGenerator string
	var crackType int

	crackJson, err := ioutil.ReadFile(*crackInfoPath)
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}

	var crack StructCrack
	err = json.Unmarshal(crackJson, &crack)
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}

	/* Process crack */
	// Check cracker
	if crack.Cracker != "" {
		crackerPath = getPath(_PATH_VENDOR) + _VENDOR_TYPE_CRACKER + PATH_SEPARATOR + crack.Cracker + PATH_SEPARATOR + task.Platform + PATH_SEPARATOR
		err := checkDir(crackerPath)
		if err != nil {
			Log.Printf("%s\n", err)
			return false
		}
		crackerPath += _VENDOR_TYPE_CRACKER
		if checkVendor(_VENDOR_TYPE_CRACKER, &crack.Cracker, &task.Platform, &crackerPath) == false {
			return false
		}
	}

	// Check generator & specify crack type
	if crack.Generator == "" {
		crackType = _CRACK_TYPE_EMBED
	} else {
		generatorPath = getPath(_PATH_VENDOR) + _VENDOR_TYPE_GENERATOR + PATH_SEPARATOR + crack.Generator + PATH_SEPARATOR + task.Platform + PATH_SEPARATOR
		err := checkDir(generatorPath)
		if err != nil {
			Log.Printf("%s\n", err)
			return false
		}
		generatorPath += _VENDOR_TYPE_GENERATOR
		if checkVendor(_VENDOR_TYPE_GENERATOR, &crack.Generator, &task.Platform, &generatorPath) == false {
			return false
		}

		if crack.Cmd_generator == "" {
			crackType = _CRACK_TYPE_STDIN
		} else {
			crackType = _CRACK_TYPE_INFILE
		}
	}

	// Check hashfile
	*crackInfoPath = filepath.Dir(*crackInfoPath) + PATH_SEPARATOR + "hashfile"
	if _, err := os.Stat(*crackInfoPath); err != nil {
		if os.IsNotExist(err) { // Does not exist, so create it
			err = ioutil.WriteFile(*crackInfoPath, []byte(crack.Target), 0664)
			if err != nil {
				Log.Printf("%s\n", err) // Error in creating
				return false
			}
		} else {
			Log.Printf("%s\n", err) // Error in accessing
			return false
		}
	}

	taskPath := getPath(_PATH_TASK) + task.Platform + PATH_SEPARATOR
	generatorReplacer := strings.NewReplacer("GENERATOR", generatorPath, "START", task.Start, "OFFSET", task.Offset, "IN_FILE", taskPath+"file.fifo")
	crackerReplacer := strings.NewReplacer("CRACKER", crackerPath, "HASH_FILE", *crackInfoPath, "OUT_FILE", taskPath+"result", "IN_FILE", taskPath+"file.fifo")

	if (crackType == _CRACK_TYPE_EMBED) || (crackType == _CRACK_TYPE_STDIN) {
		cmdCracker = generatorReplacer.Replace(crack.Cmd_cracker)
		cmdCracker = crackerReplacer.Replace(cmdCracker)

		fmt.Println("Performing crack...")

	} else if crackType == _CRACK_TYPE_INFILE {
		cmdGenerator = generatorReplacer.Replace(crack.Cmd_generator)
		cmdCracker = crackerReplacer.Replace(crack.Cmd_cracker)

		fmt.Println("Performing crack...")

		_ = cmdGenerator
	}

	return true
}
