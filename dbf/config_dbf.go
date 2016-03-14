package dbf

import (
	"encoding/json"
	"io/ioutil"
)

const _DEFAULT_URL_API = "https://d-bf.ir/api"

type StructConfDbf struct {
	Server   StructConfDbfServer     `json:"server"`
	Platform []StructConfDbfPlatform `json:"platform"`
}

type StructConfDbfServer struct {
	Url_api    string `json:"url_api"`
	Version    string `json:"version"`
	Ssl_verify int    `json:"ssl_verify"`
}

type StructConfDbfPlatform struct {
	Id        string `json:"id"`
	Active    int    `json:"active"`
	Benchmark uint64 `json:"benchmark"`
}

func createConfDbf() error {
	confDbf = StructConfDbf{
		Server: StructConfDbfServer{
			Url_api:    _DEFAULT_URL_API,
			Version:    "v1",
			Ssl_verify: 0,
		},
	}

	createPlatform()

	return saveConfDbf()
}

func saveConfDbf() error {
	confDbfJson, err := json.MarshalIndent(&confDbf, "", "\t")
	if err == nil {
		err = ioutil.WriteFile(getPath(_PATH_CONF_FILE), confDbfJson, 0664)
		return err
	} else {
		return err
	}
}

func setConfDbf() {
	confDbfJson, err := ioutil.ReadFile(getPath(_PATH_CONF_FILE))
	if err != nil {
		Log.Printf("%s\n", err)
		panic(1)
	}

	err = json.Unmarshal(confDbfJson, &confDbf)
	if err != nil {
		Log.Printf("%s\n", err)
		panic(1)
	}
}
