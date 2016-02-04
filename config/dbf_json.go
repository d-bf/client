package config

import (
	"encoding/json"
	"io/ioutil"
)

type DbfConf struct {
	Server   dbfConfServer     `json:"server"`
	Platform []dbfConfPlatform `json:"platform"`
}

type dbfConfServer struct {
	Url_api    string `json:"url_api"`
	Version    string `json:"version"`
	Ssl_verify int    `json:"ssl_verify"`
}

type dbfConfPlatform struct {
	Id        string `json:"id"`
	Active    int    `json:"active"`
	Benchmark int    `json:"benchmark"`
}

func createDbfConf() error {
	dbfConf := DbfConf{
		Server: dbfConfServer{
			Url_api:    "",
			Version:    "v1",
			Ssl_verify: 0,
		},
		Platform: createPlatforms(),
	}

	jsonDbfConf, err := json.MarshalIndent(dbfConf, "", "\t")
	if err == nil {
		err = ioutil.WriteFile(confPath, jsonDbfConf, 0644)
		return err
	} else {
		return err
	}
}
