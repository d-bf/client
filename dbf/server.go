package dbf

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

var (
	client    *http.Client
	serverUrl string
)

func initServer() {
	client = &http.Client{
		Transport: &http.Transport{
			DisableCompression: false,
		},
	}

	serverUrl = confDbf.Server.Url_api + "/" + confDbf.Server.Version + "/"
}

func setDefaultHeader(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
}

func getVendor(vendorType string, vendorName *string, platformId *string, vendorPath *string) bool {
	reqJsonStr := "{\"vendor_type\":\"" + vendorType + "\",\"name\":\"" + *vendorName + "\",\"platform_id\":\"" + *platformId + "\"}"
	reqJsonByte, err := json.Marshal(reqJsonStr)
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}

	req, err := http.NewRequest("POST", serverUrl+_URL_GET_VENDOR, bytes.NewBuffer(reqJsonByte))
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}

	setDefaultHeader(req)
	req.Header.Set("Accept", "application/octet-stream")

	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}

	if resp.StatusCode != http.StatusOK {
		Log.Printf("Bad response from server:\nStatus: %s\n Headers: %s\n", resp.Status, resp.Header)
		return false		
	}

	vendorFile, err := os.OpenFile(*vendorPath+".tmp", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0774)
	defer vendorFile.Close()
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}

	_, err = io.Copy(vendorFile, resp.Body)
	if err != nil {
		Log.Printf("%s\n", err)
		return false
	}

	os.Rename(*vendorPath+".tmp", *vendorPath)

	return true
}
