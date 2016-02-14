package dbf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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
	var crackerArg []string
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
	generatorReplacer := strings.NewReplacer("START", task.Start, "OFFSET", task.Offset, "IN_FILE", taskPath+"file.fifo")
	crackerReplacer := strings.NewReplacer("HASH_FILE", *crackInfoPath, "OUT_FILE", taskPath+"result", "IN_FILE", taskPath+"file.fifo")

	if crackType == _CRACK_TYPE_EMBED {
		cmdCracker = generatorReplacer.Replace(crack.Cmd_cracker)
		cmdCracker = crackerReplacer.Replace(cmdCracker)

		fmt.Printf("Performing crack #%s...\n", task.Crack_id)

		err = json.Unmarshal([]byte(cmdCracker), &crackerArg)
		if err != nil {
			Log.Printf("%s\n", err)
			return false
		}
		err = exec.Command(crackerPath, crackerArg...).Run()
		if err != nil {
			Log.Printf("%s\n", err)
			return false
		}

		resultByte, err := ioutil.ReadFile(taskPath + "result")
		if err != nil {
			Log.Printf("%s\n", err)
			resultByte = nil
			//			status := -1
		}

		fmt.Printf("Sending result of crack #%s...\n", task.Crack_id)

		if sendResult(`[{"crack_id":"`+task.Crack_id+`","start":"`+task.Start+`","offset":"`+task.Offset+`","result":"`+string(resultByte)+`","status":"0"}]`) == true {
			fmt.Printf("Removing task info of crack #%s (%s)...\n", task.Crack_id, task.Platform)
			err = os.RemoveAll(taskPath)
			if err != nil {
				Log.Printf("%s\n", err)
			}
		}
	} else if crackType == _CRACK_TYPE_STDIN {
		cmdGenerator = generatorReplacer.Replace(crack.Cmd_generator)
		cmdCracker = crackerReplacer.Replace(crack.Cmd_cracker)

		fmt.Printf("Performing crack #%s...\n", task.Crack_id)

		_ = cmdGenerator

	} else if crackType == _CRACK_TYPE_INFILE {
		cmdGenerator = generatorReplacer.Replace(crack.Cmd_generator)
		cmdCracker = crackerReplacer.Replace(crack.Cmd_cracker)

		fmt.Printf("Performing crack #%s...\n", task.Crack_id)

		_ = cmdGenerator
	}

	return true
}
