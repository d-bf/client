package dbf

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

var (
	client    http.Client
	serverUrl string
)

func initServer() {
	client = http.Client{
		Transport: &http.Transport{
			DisableCompression: false,
			TLSClientConfig:    &tls.Config{InsecureSkipVerify: (confDbf.Server.Ssl_verify == 0)},
		},
	}

	if validateUrl() == false {
		Log.Printf("Server URL is not valid! Please enter a valid URL in url_api in config file: %s\n", getPath(_PATH_CONF_FILE))
		panic(1)
	}

	serverUrl = confDbf.Server.Url_api + "/" + confDbf.Server.Version + "/"
}

func setDefaultHeader(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
}

func validateUrl() bool {
	req, err := http.NewRequest("POST", confDbf.Server.Url_api, bytes.NewBufferString(`{"ping":"name"}`))
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
	var pingResponse string
	err = json.NewDecoder(resp.Body).Decode(&pingResponse)
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}

	if pingResponse == "d-bf" {
		return true
	} else {
		return false
	}
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
	vendorDirPath := filepath.Dir(*vendorPath)
	err = checkDir(vendorDirPath)
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}
	os.RemoveAll(vendorDirPath) // Remove last folder

	downloadFile, err := os.OpenFile(vendorDirPath+".zip.tmp", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0774)
	if err != nil {
		Log.Printf("%s\n", err)
		downloadFile.Close()
		return false
	}

	_, err = io.Copy(downloadFile, resp.Body)
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}
	downloadFile.Close()

	err = os.Rename(vendorDirPath+".zip.tmp", vendorDirPath+".zip")
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}

	err = Uncompress(vendorDirPath+".zip", vendorDirPath+".tmp")
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}
	err = os.RemoveAll(vendorDirPath + ".zip")
	if err != nil {
		Log.Printf("%s\n", err)
	}

	err = os.Rename(vendorDirPath+".tmp", vendorDirPath)
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}

	return true
}

func GetTask() bool {
	var respTask []StructCrackTask

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
	var crack StructCrack

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

	crackJson, err := ioutil.ReadFile(*crackInfoPath + ".tmp")
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}

	err = json.Unmarshal(crackJson, &crack)
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}

	if crack.Has_dep {
		if GetCrackDep(`{"id":"`+crack.Id+`"}`, crackInfoPath) == false {
			return false
		}
	}

	err = os.Rename(*crackInfoPath+".tmp", *crackInfoPath)
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}

	return true
}

func GetCrackDep(reqJson string, crackInfoPath *string) bool {
	req, err := http.NewRequest("POST", serverUrl+_URL_GET_CRACK_DEP, bytes.NewBufferString(reqJson))
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
	crackDepPath := filepath.Dir(*crackInfoPath) + PATH_SEPARATOR + "dep"

	downloadFile, err := os.OpenFile(crackDepPath+".zip.tmp", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0774)
	if err != nil {
		Log.Printf("%s\n", err)
		downloadFile.Close()
		return false
	}

	_, err = io.Copy(downloadFile, resp.Body)
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}
	downloadFile.Close()

	err = os.Rename(crackDepPath+".zip.tmp", crackDepPath+".zip")
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}

	err = Uncompress(crackDepPath+".zip", crackDepPath+".tmp")
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}
	err = os.RemoveAll(crackDepPath + ".zip")
	if err != nil {
		Log.Printf("%s\n", err)
	}

	err = os.Rename(crackDepPath+".tmp", crackDepPath)
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}

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
