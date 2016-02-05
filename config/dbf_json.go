package config

import (
	"encoding/json"
	"github.com/d-bf/client/dbf"
	"io/ioutil"
)

type DbfConf struct {
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

func createDbfConf() error {
	dbfConf := DbfConf{
		Server: &DbfConfServer{
			Url_api:    "",
			Version:    "v1",
			Ssl_verify: 0,
		},
		Platform: createPlatforms(),
	}

	jsonDbfConf, err := json.MarshalIndent(&dbfConf, "", "\t")
	if err == nil {
		err = ioutil.WriteFile(pathConfFile, jsonDbfConf, 0644)
		return err
	} else {
		return err
	}
}

func readDbfConf() *DbfConf {
	jsonConfig, err := ioutil.ReadFile(pathConfFile)
	if err != nil {
		dbf.Log.Printf("%s\n", err)
		panic(1)
	}

	var dbfConf DbfConf

	err = json.Unmarshal(jsonConfig, &dbfConf)
	if err != nil {
		dbf.Log.Printf("%s\n", err)
		panic(1)
	}

	return &dbfConf
}
