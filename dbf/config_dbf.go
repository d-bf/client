package dbf

import (
	"encoding/json"
	"io/ioutil"
)

type ConfigDbf struct {
	Server   *ConfDbfServer     `json:"server"`
	Platform *[]ConfDbfPlatform `json:"platform"`
}

type ConfDbfServer struct {
	Url_api    string `json:"url_api"`
	Version    string `json:"version"`
	Ssl_verify int    `json:"ssl_verify"`
}

type ConfDbfPlatform struct {
	Id        string `json:"id"`
	Active    int    `json:"active"`
	Benchmark uint64 `json:"benchmark"`
}

func createConfDbf() error {
	confDbf := ConfigDbf{
		Server: &ConfDbfServer{
			Url_api:    "",
			Version:    "v1",
			Ssl_verify: 0,
		},
		Platform: createPlatform(),
	}

	return saveConfDbf(&confDbf)
}

func saveConfDbf(confDbf *ConfigDbf) error {
	confDbfJson, err := json.MarshalIndent(confDbf, "", "\t")
	if err == nil {
		err = ioutil.WriteFile(pathConfFile, confDbfJson, 0664)
		return err
	} else {
		return err
	}
}

func readConfDbf() *ConfigDbf {
	confDbfJson, err := ioutil.ReadFile(pathConfFile)
	if err != nil {
		Log.Printf("%s\n", err)
		panic(1)
	}

	var confDbf ConfigDbf

	err = json.Unmarshal(confDbfJson, &confDbf)
	if err != nil {
		Log.Printf("%s\n", err)
		panic(1)
	}

	return &confDbf
}
