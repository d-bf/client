package dbf

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

var (
	client    http.Client
	respTask  []StructCrackTask
	serverUrl string
)

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

func getVendorInfo(vendorType *string, vendorName *string, platformId *string, vendorInfoPath *string) bool {
	reqJson := `{"vendor_type":"` + *vendorType + `","name":"` + *vendorName + `","platform_id":"` + *platformId + `"}`

	req, err := http.NewRequest("POST", serverUrl+_URL_GET_VENDOR_INFO, bytes.NewBufferString(reqJson))
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
	vendorInfoFile, err := os.OpenFile(*vendorInfoPath+".tmp", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0664)
	if err != nil {
		Log.Printf("%s\n", err)
		vendorInfoFile.Close()
		return false
	}

	_, err = io.Copy(vendorInfoFile, resp.Body)
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}

	vendorInfoFile.Close()

	os.Rename(*vendorInfoPath+".tmp", *vendorInfoPath)

	return true
}

func getVendor(vendorType *string, vendorName *string, platformId *string, vendorPath *string) bool {
	reqJson := `{"vendor_type":"` + *vendorType + `","name":"` + *vendorName + `","platform_id":"` + *platformId + `"}`

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
	vendorFile, err := os.OpenFile(*vendorPath+".tmp", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0774)
	if err != nil {
		Log.Printf("%s\n", err)
		vendorFile.Close()
		return false
	}

	_, err = io.Copy(vendorFile, resp.Body)
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}

	vendorFile.Close()

	err = os.Rename(*vendorPath+".tmp", *vendorPath)
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}

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

	saveTask(&respTask)

	return true
}

func getCrackInfo(reqJson string, crackInfoPath *string) bool {
	req, err := http.NewRequest("POST", serverUrl+_URL_GET_CRACK_INFO, bytes.NewBufferString(reqJson))
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
	crackInfoFile, err := os.OpenFile(*crackInfoPath+".tmp", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0664)
	if err != nil {
		Log.Printf("%s\n", err)
		crackInfoFile.Close()
		return false
	}

	_, err = io.Copy(crackInfoFile, resp.Body)
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}

	crackInfoFile.Close()

	os.Rename(*crackInfoPath+".tmp", *crackInfoPath)

	return true
}

func sendResult(reqJson string) bool {
	req, err := http.NewRequest("POST", serverUrl+_URL_SEND_RESULT, bytes.NewBufferString(reqJson))
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

	return true
}
