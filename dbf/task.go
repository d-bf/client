package dbf

import (
	"encoding/json"
	"io/ioutil"
)

type CrackTask struct {
	Crack_id string `json:"crack_id"`
	Start    string `json:"start"`
	Offset   string `json:"offset"`
	Platform string `json:"platform"`
}

func saveTask(tasks *[]CrackTask) {
	pathTask := getPath(_PATH_TASK)
	for _, task := range *tasks {
		taskJson, err := json.MarshalIndent(&task, "", "\t")
		if err == nil {
			err = ioutil.WriteFile(pathTask+task.Platform, taskJson, 0664)
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

func processTask(task *CrackTask) {
	//pathCrack := getPath(_PATH_CRACK) + task.Crack_id + PATH_SEPARATOR + "info.json"
}
