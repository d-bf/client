package dbf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

type StructCrackTask struct {
	Crack_id string `json:"crack_id"`
	Start    string `json:"start"`
	Offset   string `json:"offset"`
	Platform string `json:"platform"`
}

var wg sync.WaitGroup

func saveTask(tasks *[]StructCrackTask) {
	taskPath := getPath(_PATH_TASK)
	for _, task := range *tasks {
		taskJson, err := json.Marshal(&task)
		if err == nil {
			err = checkDir(taskPath + task.Platform)
			if err == nil {
				err = ioutil.WriteFile(taskPath+task.Platform+PATH_SEPARATOR+"task.json", taskJson, 0664)
				if err == nil {
					wg.Add(1)
					go processTask(task)
				} else {
					Log.Printf("%s\n", err)
				}
			} else {
				Log.Printf("%s\n", err)
			}
		} else {
			Log.Printf("%s\n", err)
		}
	}

	wg.Wait() // Wait for all tasks to finish
}

func processTask(task StructCrackTask) (status bool) {
	defer func() {
		if status == false {
			Log.Printf("Error in processing crack #%s \n", task.Crack_id)
		}

		wg.Done()
	}()

	crackInfoPath := getPath(_PATH_CRACK) + task.Crack_id + PATH_SEPARATOR
	err := checkDir(crackInfoPath)
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}
	crackInfoPath += "crack.json"

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
	return processCrack(&task, &crackInfoPath)
}
