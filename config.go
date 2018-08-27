package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type tvnetInfo struct {
	Tv      tvInfo
	Network networkInfo
}

type televisions *[]tvnetInfo

var (
	configHome           = fmt.Sprintf("%s/.config/smarttv-remote", os.Getenv("HOME"))
	televisionConfigFile = fmt.Sprintf("%s/tvsConfig.json", configHome)
)

func readTelevisionConfig() (config televisions, e error) {
	var (
		fh *os.File
		b  []byte
	)

	if fh, e = os.Open(televisionConfigFile); e != nil {
		return
	}
	defer fh.Close()

	if b, e = ioutil.ReadAll(fh); e != nil {
		return
	}

	config = new([]tvnetInfo)
	e = json.Unmarshal(b, config)

	return
}

func writeTelevisionConfig(tvs televisions) (e error) {
	var (
		fh *os.File
		b  []byte
	)

	e = os.MkdirAll(configHome, 0770)

	flags := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if fh, e = os.OpenFile(televisionConfigFile, flags, 0770); e != nil {
		return
	}
	defer fh.Close()

	if b, e = json.Marshal(tvs); e != nil {
		return
	}

	_, e = fh.Write(b)

	return
}
