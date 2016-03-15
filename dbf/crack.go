package dbf

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const (
	_CRACK_TYPE_INFILE = 1
	_CRACK_TYPE_STDIN  = 2
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

type StructNotEmbed struct {
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

func processCrack(task StructCrackTask) (status bool) {
	var jsonByte []byte
	var vendorPath, cmdJsonStr string
	var cmdArg []string
	var resultStatus int

	resultStatus = -1
	taskPath := getPath(_PATH_TASK) + task.Platform + PATH_SEPARATOR

	defer func() {
		resultByte, err := ioutil.ReadFile(taskPath + "result")
		resultStr := ""
		if err == nil {
			resultStr = base64.StdEncoding.EncodeToString(resultByte)
		} else {
			if !os.IsNotExist(err) {
				Log.Printf("%s\n", err)
				resultStatus = -2
			}
		}

		fmt.Printf("Sending result of crack #%s (%s, status: %d)...\n", task.Crack_id, task.Platform, resultStatus)

		if sendResult(`[{"crack_id":"`+task.Crack_id+`","start":"`+task.Start+`","offset":"`+task.Offset+`","result":"`+resultStr+`","status":"`+strconv.Itoa(resultStatus)+`"}]`) == true {
			fmt.Printf("Remove current task info of crack #%s (%s)...\n", task.Crack_id, task.Platform)
			err = os.RemoveAll(taskPath)
			if err != nil {
				Log.Printf("%s\n", err)
			}
		} else {
			Log.Printf("Error in sending result of crack #%s (%s)\n", task.Crack_id, task.Platform)
		}

		if status {
			ResetTimer = true
		} else {
			Log.Printf("Error in processing crack #%s (%s)\n", task.Crack_id, task.Platform)
		}

		wgTask.Done()
	}()

	crackInfoPath := getPath(_PATH_CRACK) + task.Crack_id + PATH_SEPARATOR + "crack.json"
	crackJson, err := ioutil.ReadFile(crackInfoPath)
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
	crackInfoPath = filepath.Dir(crackInfoPath) + PATH_SEPARATOR + "hashfile"
	if _, err := os.Stat(crackInfoPath); err != nil {
		if os.IsNotExist(err) { // Does not exist, so create it
			err = ioutil.WriteFile(crackInfoPath, []byte(crack.Target), 0664)
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
	crackerReplacer := strings.NewReplacer("ALGO_ID", crack.Algo_id, "ALGO_NAME", crack.Algo_name, `"HASH_FILE"`, strconv.Quote(crackInfoPath), `"OUT_FILE"`, strconv.Quote(taskPath+"result"), `"IN_FILE"`, strconv.Quote(taskPath+"file.fifo"))

	if internal_gen { // Embeded
		// Get cracker info
		vendorPath = filepath.Dir(vendorPath) + PATH_SEPARATOR + "info.json"
		jsonByte, err = ioutil.ReadFile(vendorPath)
		if err != nil {
			Log.Printf("%s\n", err)
			resultStatus = -9
			return false
		}

		var crackerEmbed StructCrackerEmbed
		err = json.Unmarshal(jsonByte, &crackerEmbed)
		if err != nil {
			Log.Printf("%s\n", err)
			resultStatus = -10
			return false
		}

		// Get arg
		brk := false
		for _, crackerGen := range crackerEmbed.Generator {
			if crackerGen.Name == crack.Generator {
				cmdArg = crackerGen.Arg
				brk = true
				break
			}
		}
		if brk == false {
			Log.Printf("No embedded arg for generator '%s' of cracker '%s' in info.json!\n", crack.Generator, cracker)
			resultStatus = -11
			return false
		}

		// Replace arg
		jsonByte, err = json.Marshal(&cmdArg)
		if err != nil {
			Log.Printf("%s\n", err)
			resultStatus = -12
			return false
		}
		cmdJsonStr = string(jsonByte)
		cmdJsonStr = generatorReplacer.Replace(cmdJsonStr)
		cmdJsonStr = crackerReplacer.Replace(cmdJsonStr)
		err = json.Unmarshal([]byte(cmdJsonStr), &cmdArg)
		if err != nil {
			Log.Printf("%s\n", err)
			resultStatus = -13
			return false
		}

		fmt.Printf("Performing crack #%s (%s)...\n", task.Crack_id, task.Platform)

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
	} else { // Not embeded
		var crackType int

		/* Determine stdin or infile */
		// Get cracker info
		vendorPath = filepath.Dir(vendorPath) + PATH_SEPARATOR + "info.json"
		jsonByte, err = ioutil.ReadFile(vendorPath)
		if err != nil {
			Log.Printf("%s\n", err)
			resultStatus = -15
			return false
		}

		var crackerNEmbed StructNotEmbed
		err = json.Unmarshal(jsonByte, &crackerNEmbed)
		if err != nil {
			Log.Printf("%s\n", err)
			resultStatus = -16
			return false
		}

		// Check generator
		vendorPath = getPath(_PATH_VENDOR) + _VENDOR_TYPE_GENERATOR + PATH_SEPARATOR + crack.Generator + PATH_SEPARATOR + task.Platform + PATH_SEPARATOR + _VENDOR_TYPE_GENERATOR + extExecutable
		if checkVendor(_VENDOR_TYPE_GENERATOR, &crack.Generator, &task.Platform, &vendorPath) == false {
			resultStatus = -17
			return false
		}

		// Get generator info
		vendorPath = filepath.Dir(vendorPath) + PATH_SEPARATOR + "info.json"
		jsonByte, err = ioutil.ReadFile(vendorPath)
		if err != nil {
			Log.Printf("%s\n", err)
			resultStatus = -18
			return false
		}

		var generatorNEmbed StructNotEmbed
		err = json.Unmarshal(jsonByte, &generatorNEmbed)
		if err != nil {
			Log.Printf("%s\n", err)
			resultStatus = -19
			return false
		}

		// First check for stdin then infile
		if (len(crackerNEmbed.Stdin) > 0) && (len(generatorNEmbed.Stdin) > 0) {
			crackType = _CRACK_TYPE_STDIN
		} else if (len(crackerNEmbed.Infile) > 0) && (len(generatorNEmbed.Infile) > 0) {
			crackType = _CRACK_TYPE_INFILE
		} else {
			Log.Printf("Can't find stdin nor infile config for generator '%s' and cracker '%s'\n", crack.Generator, cracker)
			resultStatus = -20
			return false
		}

		/* Prepare cracker */
		// Set cracker arg
		if crackType == _CRACK_TYPE_STDIN {
			cmdArg = crackerNEmbed.Stdin
		} else {
			cmdArg = crackerNEmbed.Infile
		}

		// Replace cracker arg
		jsonByte, err = json.Marshal(&cmdArg)
		if err != nil {
			Log.Printf("%s\n", err)
			resultStatus = -21
			return false
		}
		cmdJsonStr = string(jsonByte)
		cmdJsonStr = crackerReplacer.Replace(cmdJsonStr)
		err = json.Unmarshal([]byte(cmdJsonStr), &cmdArg)
		if err != nil {
			Log.Printf("%s\n", err)
			resultStatus = -22
			return false
		}

		execCracker := exec.Command(getPath(_PATH_VENDOR)+_VENDOR_TYPE_CRACKER+PATH_SEPARATOR+cracker+PATH_SEPARATOR+task.Platform+PATH_SEPARATOR+_VENDOR_TYPE_CRACKER+extExecutable, cmdArg...)

		/* Prepare generator */
		// Set generator arg
		if crackType == _CRACK_TYPE_STDIN {
			cmdArg = generatorNEmbed.Stdin
		} else {
			cmdArg = generatorNEmbed.Infile
		}

		// Replace generator arg
		jsonByte, err = json.Marshal(&cmdArg)
		if err != nil {
			Log.Printf("%s\n", err)
			resultStatus = -23
			return false
		}
		cmdJsonStr = string(jsonByte)
		cmdJsonStr = generatorReplacer.Replace(cmdJsonStr)
		if strings.Contains(cmdJsonStr, "DEP_GEN") {
			// Check if dependency exists in crack location
			crackInfoPath = filepath.Dir(crackInfoPath) + PATH_SEPARATOR + "dep" + PATH_SEPARATOR + "dep-gen"
			if _, err := os.Stat(crackInfoPath); err == nil { // dep-gen file exists in crack location and is accessible
				cmdJsonStr = strings.Replace(cmdJsonStr, `"DEP_GEN"`, strconv.Quote(crackInfoPath), -1)
			} else { // Check if dependency exists in generator location
				vendorPath = filepath.Dir(vendorPath) + PATH_SEPARATOR + "dep-gen"
				if _, err := os.Stat(vendorPath); err == nil { // dep-gen file exists in generator location and is accessible
					cmdJsonStr = strings.Replace(cmdJsonStr, `"DEP_GEN"`, strconv.Quote(vendorPath), -1)
				} else {
					Log.Printf("Dependency not found! crack #%d, generator: '%s'\n", crack.Id, crack.Generator)
					resultStatus = -24
					return false
				}
			}
		}
		err = json.Unmarshal([]byte(cmdJsonStr), &cmdArg)
		if err != nil {
			Log.Printf("%s\n", err)
			resultStatus = -25
			return false
		}

		vendorPath = filepath.Dir(vendorPath) + PATH_SEPARATOR + _VENDOR_TYPE_GENERATOR + extExecutable
		execGenerator := exec.Command(vendorPath, cmdArg...)

		fmt.Printf("Performing crack #%s (%s)...\n", task.Crack_id, task.Platform)

		if crackType == _CRACK_TYPE_STDIN {
			r, err := execGenerator.StdoutPipe()
			if err != nil {
				Log.Printf("%s\n", err)
				resultStatus = -26
				return false
			}
			execCracker.Stdin = r

			err = execGenerator.Start()
			if err != nil {
				Log.Printf("%s\n", err)
				resultStatus = -27
				return false
			}

			err = execCracker.Start()
			if err != nil {
				Log.Printf("%s\n", err)
				resultStatus = -28
				return false
			}

			execCracker.Wait()
			execGenerator.Process.Signal(syscall.SIGINT) // ^C (Control-C)
			r.Close()

			resultStatus = 0
			return true
		} else { // Infile
			err = exec.Command("mkfifo", taskPath+"file.fifo").Run()
			if err != nil {
				Log.Printf("%s\n", err)
				resultStatus = -29
				return false
			}

			err = execGenerator.Start()
			if err != nil {
				Log.Printf("%s\n", err)
				resultStatus = -30
				return false
			}

			err = execCracker.Start()
			if err != nil {
				Log.Printf("%s\n", err)
				resultStatus = -31
				return false
			}

			execCracker.Wait()
			execGenerator.Process.Signal(syscall.SIGINT) // ^C (Control-C)

			resultStatus = 0
			return true
		}
	}

	return true
}
