package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/d-bf/client/config"
	"github.com/d-bf/client/dbf"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

var (
	client    *http.Client
	serverUrl string
)

func Init() {
	client = &http.Client{
		Transport: &http.Transport{
			DisableCompression: false,
		},
	}

	serverUrl = config.DbfConfig.Server.Url_api + "/" + config.DbfConfig.Server.Version + "/"
}

func setDefaultHeader(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
}

func GetVendor(vendorType string, vendorName *string, platform *string) bool {
	reqJsonStr := "{\"vendor_type\":\"" + vendorType + "\",\"name\":\"" + *vendorName + "\",\"platform_id\":\"" + *platform + "\"}"
	reqJsonByte, err := json.Marshal(reqJsonStr)
	if err != nil {
		dbf.Log.Printf("%s\n", err)
		return false
	}

	req, err := http.NewRequest("POST", serverUrl+"vendor/get", bytes.NewBuffer(reqJsonByte))
	if err != nil {
		dbf.Log.Printf("%s\n", err)
		return false
	}

	setDefaultHeader(req)

	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		dbf.Log.Printf("%s\n", err)
		return false
	}

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)

	file, err := os.Create("")
	defer file.Close()
	if err != nil {
		dbf.Log.Printf("%s\n", err)
		return false
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		dbf.Log.Printf("%s\n", err)
		return false
	}

	return true
}

func Test() {
	req, _ := http.NewRequest("GET", serverUrl, nil)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		fmt.Println("Error")
	} else {
		fmt.Println("response Status:", resp.Status)
		fmt.Println("response Headers:", resp.Header)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("response Body:", string(body))
	}
}
