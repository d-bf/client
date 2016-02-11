package dbf

import (
	"encoding/json"
	"io/ioutil"
)

type Crack struct {
	Id            string `json:"id"`
	Generator     string `json:"generator"`
	Cracker       string `json:"cracker"`
	Cmd_generator string `json:"cmd_generator"`
	Cmd_cracker   string `json:"cmd_cracker"`
}

func processCrack(task *CrackTask, crackInfoPath *string) bool {
	crackJson, err := ioutil.ReadFile(*crackInfoPath)
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}

	var crack Crack
	err = json.Unmarshal(crackJson, &crack)
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}

	return true
}
