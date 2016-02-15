package dbf

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type StructCrack struct {
	Id            string `json:"id"`
	Type          string `json:"type"`
	Generator     string `json:"generator"`
	Cracker       string `json:"cracker"`
	Cmd_generator string `json:"cmd_generator"`
	Cmd_cracker   string `json:"cmd_cracker"`
	Target        string `json:"target"`
}

func processCrack(task *StructCrackTask, crackInfoPath *string) bool {
	var vendorPath, cmdJsonStr string
	var cmdArg []string
	var resultStatus int

	resultStatus = -1
	taskPath := getPath(_PATH_TASK) + task.Platform + PATH_SEPARATOR

	defer func() {
		resultByte, err := ioutil.ReadFile(taskPath + "result")
		if err != nil {
			Log.Printf("%s\n", err)
			resultByte = nil
			resultStatus = -2
		}

		fmt.Printf("Sending result of crack #%s...\n", task.Crack_id)

		if sendResult(`[{"crack_id":"`+task.Crack_id+`","start":"`+task.Start+`","offset":"`+task.Offset+`","result":"`+string(resultByte)+`","status":"`+strconv.Itoa(resultStatus)+`"}]`) == true {
			fmt.Printf("Removing task info of crack #%s (%s)...\n", task.Crack_id, task.Platform)
			err = os.RemoveAll(taskPath)
			if err != nil {
				Log.Printf("%s\n", err)
			}
		}
	}()

	crackJson, err := ioutil.ReadFile(*crackInfoPath)
	if err != nil {
		Log.Printf("%s\n", err)
		resultStatus = -3
		return false
	}

	var crack StructCrack
	err = json.Unmarshal(crackJson, &crack)
	if err != nil {
		Log.Printf("%s\n", err)
		resultStatus = -4
		return false
	}

	/* Process crack */
	// Check cracker
	if crack.Cracker != "" {
		vendorPath = getPath(_PATH_VENDOR) + _VENDOR_TYPE_CRACKER + PATH_SEPARATOR + crack.Cracker + PATH_SEPARATOR + task.Platform + PATH_SEPARATOR
		err := checkDir(vendorPath)
		if err != nil {
			Log.Printf("%s\n", err)
			resultStatus = -5
			return false
		}
		vendorPath += _VENDOR_TYPE_CRACKER + extExecutable
		if checkVendor(_VENDOR_TYPE_CRACKER, &crack.Cracker, &task.Platform, &vendorPath) == false {
			resultStatus = -6
			return false
		}
	}

	// Check hashfile
	*crackInfoPath = filepath.Dir(*crackInfoPath) + PATH_SEPARATOR + "hashfile"
	if _, err := os.Stat(*crackInfoPath); err != nil {
		if os.IsNotExist(err) { // Does not exist, so create it
			err = ioutil.WriteFile(*crackInfoPath, []byte(crack.Target), 0664)
			if err != nil {
				Log.Printf("%s\n", err) // Error in creating
				resultStatus = -7
				return false
			}
		} else {
			Log.Printf("%s\n", err) // Error in accessing
			resultStatus = -8
			return false
		}
	}

	generatorReplacer := strings.NewReplacer("START", task.Start, "OFFSET", task.Offset, "IN_FILE", taskPath+"file.fifo")
	crackerReplacer := strings.NewReplacer("HASH_FILE", *crackInfoPath, "OUT_FILE", taskPath+"result", "IN_FILE", taskPath+"file.fifo")

	if crack.Type == "embed" { // Embeded
		cmdJsonStr = generatorReplacer.Replace(crack.Cmd_cracker)
		cmdJsonStr = crackerReplacer.Replace(cmdJsonStr)
		err = json.Unmarshal([]byte(cmdJsonStr), &cmdArg)
		if err != nil {
			Log.Printf("%s\n", err)
			resultStatus = -9
			return false
		}

		fmt.Printf("Performing crack #%s...\n", task.Crack_id)

		err = exec.Command(vendorPath, cmdArg...).Run()
		if err != nil {
			Log.Printf("%s\n", err)
			resultStatus = -10
			return false
		} else {
			resultStatus = 0
			return true
		}
	} else { // Not embeded
		// Prepare cracker
		cmdJsonStr = crackerReplacer.Replace(crack.Cmd_cracker)
		err = json.Unmarshal([]byte(cmdJsonStr), &cmdArg)
		if err != nil {
			Log.Printf("%s\n", err)
			resultStatus = -11
			return false
		}
		execCracker := exec.Command(vendorPath, cmdArg...)

		// Check generator
		vendorPath = getPath(_PATH_VENDOR) + _VENDOR_TYPE_GENERATOR + PATH_SEPARATOR + crack.Generator + PATH_SEPARATOR + task.Platform + PATH_SEPARATOR
		err := checkDir(vendorPath)
		if err != nil {
			Log.Printf("%s\n", err)
			resultStatus = -12
			return false
		}
		vendorPath += _VENDOR_TYPE_GENERATOR + extExecutable
		if checkVendor(_VENDOR_TYPE_GENERATOR, &crack.Generator, &task.Platform, &vendorPath) == false {
			resultStatus = -13
			return false
		}

		// Prepare generator
		cmdJsonStr = generatorReplacer.Replace(crack.Cmd_generator)
		err = json.Unmarshal([]byte(cmdJsonStr), &cmdArg)
		if err != nil {
			Log.Printf("%s\n", err)
			resultStatus = -14
			return false
		}
		execGenerator := exec.Command(vendorPath, cmdArg...)

		fmt.Printf("Performing crack #%s...\n", task.Crack_id)

		if crack.Type == "infile" {
			err = exec.Command("mkfifo", taskPath+"file.fifo").Run()
			if err != nil {
				Log.Printf("%s\n", err)
				resultStatus = -15
				return false
			}

			err = execGenerator.Start()
			if err != nil {
				Log.Printf("%s\n", err)
				resultStatus = -16
				return false
			}

			err = execCracker.Start()
			if err != nil {
				Log.Printf("%s\n", err)
				resultStatus = -17
				return false
			}

			errG := execGenerator.Wait()
			errC := execCracker.Wait()
			if (errG != nil) || (errC != nil) {

				resultStatus = -17

				if errG != nil {
					Log.Printf("%s\n", errG)
					resultStatus += -1
				} else if errC != nil {
					Log.Printf("%s\n", errC)
					resultStatus += -2
				}

				// Max resultStatus: -20

				return false
			} else {
				resultStatus = 0
				return true
			}
		} else { // Stdin
			r, w := io.Pipe()
			execGenerator.Stdout = w
			execCracker.Stdin = r

			err = execGenerator.Start()
			if err != nil {
				Log.Printf("%s\n", err)
				resultStatus = -21
				return false
			}

			err = execCracker.Start()
			if err != nil {
				Log.Printf("%s\n", err)
				resultStatus = -22
				return false
			}

			errG := execGenerator.Wait()
			errW := w.Close()
			errC := execCracker.Wait()
			if (errG != nil) || (errW != nil) || (errC != nil) {
				resultStatus = -22

				if errG != nil {
					Log.Printf("%s\n", errG)
					resultStatus += -1
				} else if errW != nil {
					Log.Printf("%s\n", errW)
					resultStatus += -2
				} else if errC != nil {
					Log.Printf("%s\n", errC)
					resultStatus += -4
				}

				// Max resultStatus: -29

				return false
			} else {
				resultStatus = 0
				return true
			}
		}
	}

	return true
}
