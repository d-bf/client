package dbf

import (
	"encoding/json"
	"io/ioutil"
)

type ConfigDbf struct {
	Server   *DbfConfServer     `json:"server"`
	Platform *[]DbfConfPlatform `json:"platform"`
}

type DbfConfServer struct {
	Url_api    string `json:"url_api"`
	Version    string `json:"version"`
	Ssl_verify int    `json:"ssl_verify"`
}

type DbfConfPlatform struct {
	Id        string `json:"id"`
	Active    int    `json:"active"`
	Benchmark int    `json:"benchmark"`
}

func createConfDbf() error {
	ConfDbf := ConfigDbf{
		Server: &DbfConfServer{
			Url_api:    "",
			Version:    "v1",
			Ssl_verify: 0,
		},
		Platform: createPlatform(),
	}

	confDbfJson, err := json.MarshalIndent(&ConfDbf, "", "\t")
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
