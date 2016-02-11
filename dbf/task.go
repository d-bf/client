package dbf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type StructCrackTask struct {
	Crack_id string `json:"crack_id"`
	Start    string `json:"start"`
	Offset   string `json:"offset"`
	Platform string `json:"platform"`
}

func saveTask(tasks *[]StructCrackTask) {
	taskPath := getPath(_PATH_TASK)
	for _, task := range *tasks {
		taskJson, err := json.Marshal(&task)
		if err == nil {
			err = ioutil.WriteFile(taskPath+task.Platform, taskJson, 0664)
			if err == nil {
				processTask(&task)
			} else {
				Log.Printf("%s\n", err)
			}
		} else {
			Log.Printf("%s\n", err)
		}
	}
}

func processTask(task *StructCrackTask) bool {
	crackInfoPath := getPath(_PATH_CRACK) + task.Crack_id + PATH_SEPARATOR
	err := checkDir(crackInfoPath)
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}
	crackInfoPath += "info.json"

	if _, err := os.Stat(crackInfoPath); err != nil {
		if os.IsNotExist(err) { // Does not exist, so get it

			fmt.Printf("Getting crack info of crack #%s...\n", task.Crack_id)

			if getCrackInfo(`{"id":"`+task.Crack_id+`","platform":"`+task.Platform+`"}`, &crackInfoPath) == false {
				return false
			}
		} else {
			Log.Printf("%s\n", err) // Error in accessing
			return false
		}
	}

	// Process crack
	return processCrack(task, &crackInfoPath)
}
