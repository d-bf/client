package dbf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	//	"syscall"
)

type StructCrack struct {
	Id         string            `json:"id"`
	Generator  string            `json:"generator"`
	Gen_config []string          `json:"gen_config"`
	Algo_id    string            `json:"algo_id"`
	Algo_name  string            `json:"algo_name"`
	Len_min    string            `json:"len_min"`
	Len_max    string            `json:"len_max"`
	Charset1   string            `json:"charset1"`
	Charset2   string            `json:"charset2"`
	Charset3   string            `json:"charset3"`
	Charset4   string            `json:"charset4"`
	Mask       string            `json:"mask"`
	Target     string            `json:"target"`
	Has_dep    bool              `json:"has_dep"`
	Info       []StructCrackInfo `json:"info"`
}

type StructCrackInfo struct {
	Platform     string `json:"platform"`
	Cracker      string `json:"cracker"`
	Internal_gen bool   `json:"internal_gen"`
}

type StructCrackerNEmbed struct {
	Stdin  []string `json:"stdin"`
	Infile []string `json:"infile"`
}

type StructCrackerEmbed struct {
	Generator []StructCrackerGen `json:"generator"`
}

type StructCrackerGen struct {
	Name string   `json:"name"`
	Arg  []string `json:"arg"`
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
			resultByte = nil

			if os.IsNotExist(err) {
				resultStatus = 0 // Was not cracked
			} else {
				Log.Printf("%s\n", err)
				resultStatus = -2
			}
		}

		fmt.Printf("Sending result of crack #%s (status: %d)...\n", task.Crack_id, resultStatus)

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
	var cracker string
	var internal_gen bool
	for _, crackInfo := range crack.Info {
		if crackInfo.Platform == task.Platform {
			cracker = crackInfo.Cracker
			internal_gen = crackInfo.Internal_gen
			break
		}
	}

	if cracker == "" {
		resultStatus = -5
		return false
	}

	vendorPath = getPath(_PATH_VENDOR) + _VENDOR_TYPE_CRACKER + PATH_SEPARATOR + cracker + PATH_SEPARATOR + task.Platform + PATH_SEPARATOR + _VENDOR_TYPE_CRACKER + extExecutable
	if checkVendor(_VENDOR_TYPE_CRACKER, &cracker, &task.Platform, &vendorPath) == false {
		resultStatus = -6
		return false
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

	/* Quote question mark in mask! */
	crack.Mask = strings.Replace(crack.Mask, "?", "??", -1)
	crack.Mask = strings.Replace(crack.Mask, "??l", "?l", -1)
	crack.Mask = strings.Replace(crack.Mask, "??u", "?u", -1)
	crack.Mask = strings.Replace(crack.Mask, "??d", "?d", -1)
	crack.Mask = strings.Replace(crack.Mask, "??s", "?s", -1)
	crack.Mask = strings.Replace(crack.Mask, "??a", "?a", -1)
	crack.Mask = strings.Replace(crack.Mask, "??b", "?b", -1)
	crack.Mask = strings.Replace(crack.Mask, "??1", "?1", -1)
	crack.Mask = strings.Replace(crack.Mask, "??2", "?2", -1)
	crack.Mask = strings.Replace(crack.Mask, "??3", "?3", -1)
	crack.Mask = strings.Replace(crack.Mask, "??4", "?4", -1)

	/* Handle replacement of custom charsets */
	var char1, char2, char3, char4 string
	if len(crack.Charset1) > 0 {
		char1 = `,"-1",` + strconv.Quote(strings.Replace(crack.Charset1, "?", "??", -1))
	} else {
		char1 = ``
	}

	if len(crack.Charset2) > 0 {
		char2 = `,"-2",` + strconv.Quote(strings.Replace(crack.Charset2, "?", "??", -1))
	} else {
		char2 = ``
	}

	if len(crack.Charset3) > 0 {
		char3 = `,"-3",` + strconv.Quote(strings.Replace(crack.Charset3, "?", "??", -1))
	} else {
		char3 = ``
	}

	if len(crack.Charset4) > 0 {
		char4 = `,"-4",` + strconv.Quote(strings.Replace(crack.Charset4, "?", "??", -1))
	} else {
		char4 = ``
	}

	generatorReplacer := strings.NewReplacer("START", task.Start, "OFFSET", task.Offset, "LEN_MIN", crack.Len_min, "LEN_MAX", crack.Len_max, `,"CHAR1"`, char1, `,"CHAR2"`, char2, `,"CHAR3"`, char3, `,"CHAR4"`, char4, "MASK", crack.Mask, `"IN_FILE"`, strconv.Quote(taskPath+"file.fifo"))
	crackerReplacer := strings.NewReplacer("ALGO_ID", crack.Algo_id, "ALGO_NAME", crack.Algo_name, `"HASH_FILE"`, strconv.Quote(*crackInfoPath), `"OUT_FILE"`, strconv.Quote(taskPath+"result"), `"IN_FILE"`, strconv.Quote(taskPath+"file.fifo"))

	if internal_gen { // Embeded
		vendorPath = filepath.Dir(vendorPath) + PATH_SEPARATOR + "info.json"
		crackerJson, err := ioutil.ReadFile(vendorPath)
		if err != nil {
			Log.Printf("%s\n", err)
			resultStatus = -9
			return false
		}

		var crackerEmbed StructCrackerEmbed
		err = json.Unmarshal(crackerJson, &crackerEmbed)
		if err != nil {
			Log.Printf("%s\n", err)
			resultStatus = -10
			return false
		}

		brk := false
		for _, crackerGen := range crackerEmbed.Generator {
			if crackerGen.Name == crack.Generator {
				cmdArg = crackerGen.Arg
				brk = true
				break
			}
		}
		if brk == false {
			Log.Printf("No args for cracker '%s' in info.json!\n", cracker)
			resultStatus = -11
			return false
		}

		cmdJsonByte, err := json.Marshal(&cmdArg)
		if err != nil {
			Log.Printf("%s\n", err)
			resultStatus = -12
			return false
		}
		cmdJsonStr = string(cmdJsonByte)

		cmdJsonStr = generatorReplacer.Replace(cmdJsonStr)
		cmdJsonStr = crackerReplacer.Replace(cmdJsonStr)
		err = json.Unmarshal([]byte(cmdJsonStr), &cmdArg)
		if err != nil {
			Log.Printf("%s\n", err)
			resultStatus = -13
			return false
		}

		fmt.Printf("Performing crack #%s...\n", task.Crack_id)

		vendorPath = filepath.Dir(vendorPath) + PATH_SEPARATOR + _VENDOR_TYPE_CRACKER + extExecutable
		err = exec.Command(vendorPath, cmdArg...).Run()
		if err != nil {
			Log.Printf("%s\n", err)
			resultStatus = -14
			return false
		} else {
			resultStatus = 0
			return true
		}
	}
	//	 else { // Not embeded
	//		// Prepare cracker
	//		cmdJsonStr = crackerReplacer.Replace(crack.Cmd_cracker)
	//		err = json.Unmarshal([]byte(cmdJsonStr), &cmdArg)
	//		if err != nil {
	//			Log.Printf("%s\n", err)
	//			resultStatus = -10
	//			return false
	//		}
	//		execCracker := exec.Command(vendorPath, cmdArg...)
	//
	//		// Check generator
	//		vendorPath = getPath(_PATH_VENDOR) + _VENDOR_TYPE_GENERATOR + PATH_SEPARATOR + crack.Generator + PATH_SEPARATOR + task.Platform + PATH_SEPARATOR + _VENDOR_TYPE_GENERATOR + extExecutable
	//		if checkVendor(_VENDOR_TYPE_GENERATOR, &crack.Generator, &task.Platform, &vendorPath) == false {
	//			resultStatus = -11
	//			return false
	//		}
	//
	//		// Prepare generator
	//		cmdJsonStr = generatorReplacer.Replace(crack.Cmd_generator)
	//		if strings.Contains(cmdJsonStr, "DEP_GEN") {
	//			// Check if dependency exists in crack location
	//			*crackInfoPath = filepath.Dir(*crackInfoPath) + PATH_SEPARATOR + "dep" + PATH_SEPARATOR + "dep-gen"
	//			if _, err := os.Stat(*crackInfoPath); err == nil { // dep-gen file exists in crack location and is accessible
	//				cmdJsonStr = strings.Replace(cmdJsonStr, `"DEP_GEN"`, strconv.Quote(*crackInfoPath), -1)
	//			} else { // Check if dependency exists in generator location
	//				vendorPath = filepath.Dir(vendorPath) + PATH_SEPARATOR + "dep-gen"
	//				if _, err := os.Stat(vendorPath); err == nil { // dep-gen file exists in generator location and is accessible
	//					cmdJsonStr = strings.Replace(cmdJsonStr, `"DEP_GEN"`, strconv.Quote(vendorPath), -1)
	//				} else {
	//					resultStatus = -12
	//					return false
	//				}
	//				vendorPath = filepath.Dir(vendorPath) + PATH_SEPARATOR + _VENDOR_TYPE_GENERATOR + extExecutable // Rename back to generator executable
	//			}
	//		}
	//		err = json.Unmarshal([]byte(cmdJsonStr), &cmdArg)
	//		if err != nil {
	//			Log.Printf("%s\n", err)
	//			resultStatus = -13
	//			return false
	//		}
	//		execGenerator := exec.Command(vendorPath, cmdArg...)
	//
	//		fmt.Printf("Performing crack #%s...\n", task.Crack_id)
	//
	//		if crack.Type == "stdin" {
	//			r, err := execGenerator.StdoutPipe()
	//			if err != nil {
	//				Log.Printf("%s\n", err)
	//				resultStatus = -14
	//				return false
	//			}
	//			execCracker.Stdin = r
	//
	//			err = execGenerator.Start()
	//			if err != nil {
	//				Log.Printf("%s\n", err)
	//				resultStatus = -15
	//				return false
	//			}
	//
	//			err = execCracker.Start()
	//			if err != nil {
	//				Log.Printf("%s\n", err)
	//				resultStatus = -16
	//				return false
	//			}
	//
	//			execCracker.Wait()
	//			execGenerator.Process.Signal(syscall.SIGINT) // ^C (Control-C)
	//			r.Close()
	//
	//			resultStatus = 0
	//			return true
	//		} else { // Infile
	//			err = exec.Command("mkfifo", taskPath+"file.fifo").Run()
	//			if err != nil {
	//				Log.Printf("%s\n", err)
	//				resultStatus = -18
	//				return false
	//			}
	//
	//			err = execGenerator.Start()
	//			if err != nil {
	//				Log.Printf("%s\n", err)
	//				resultStatus = -19
	//				return false
	//			}
	//
	//			err = execCracker.Start()
	//			if err != nil {
	//				Log.Printf("%s\n", err)
	//				resultStatus = -20
	//				return false
	//			}
	//
	//			execCracker.Wait()
	//			execGenerator.Process.Signal(syscall.SIGINT) // ^C (Control-C)
	//
	//			resultStatus = 0
	//			return true
	//		}
	//	}

	return true
}
