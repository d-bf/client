package dbf

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

var (
	client    http.Client
	respTask  []ResponseTask
	serverUrl string
)

type ResponseTask struct {
	Crack_id string `json:"crack_id"`
	Start    string `json:"start"`
	Offset   string `json:"offset"`
	Platform string `json:"platform"`
}

func initServer() {
	client = http.Client{
		Transport: &http.Transport{
			DisableCompression: false,
			TLSClientConfig:    &tls.Config{InsecureSkipVerify: (confDbf.Server.Ssl_verify == 0)},
		},
	}

	serverUrl = confDbf.Server.Url_api + "/" + confDbf.Server.Version + "/"
}

func setDefaultHeader(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
}

func getVendor(vendorType string, vendorName *string, platformId *string, vendorPath *string) bool {
	reqJson := `{"vendor_type":"` + vendorType + `","name":"` + *vendorName + `","platform_id":"` + *platformId + `"}`

	req, err := http.NewRequest("POST", serverUrl+_URL_GET_VENDOR, bytes.NewBufferString(reqJson))
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}

	setDefaultHeader(req)
	req.Header.Set("Accept", "application/octet-stream")

	resp, err := client.Do(req)
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		Log.Printf("Bad response from server:\nStatus: %s\nHeaders: %s\n", resp.Status, resp.Header)
		return false
	}

	// Process response
	vendorFile, err := os.OpenFile(*vendorPath+".tmp", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0774)
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}
	defer vendorFile.Close()

	_, err = io.Copy(vendorFile, resp.Body)
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}

	os.Rename(*vendorPath+".tmp", *vendorPath)

	return true
}

func GetTask() bool {
	reqJson := `{"client_info":{"platform":` + activePlatStr + `}}`

	req, err := http.NewRequest("POST", serverUrl+_URL_GET_TASK, bytes.NewBufferString(reqJson))
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}

	setDefaultHeader(req)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		Log.Printf("Bad response from server:\nStatus: %s\nHeaders: %s\n", resp.Status, resp.Header)
		return false
	}

	// Process response
	err = json.NewDecoder(resp.Body).Decode(&respTask)
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}

	for i, task := range respTask {
		fmt.Println("Task", i, ":", task)
	}

	return true
}
