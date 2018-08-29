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
type keymap map[byte]string

type keyInput struct {
	key   string
	mod   []string
	event string
}

var (
	configHome           = fmt.Sprintf("%s/.config/smarttv-remote", os.Getenv("HOME"))
	televisionConfigFile = fmt.Sprintf("%s/tvsConfig.json", configHome)
	keymapConfigFile     = fmt.Sprintf("%s/keymap.json", configHome)

	sdlDefaultKeymap = []keyInput{
		{"0", nil, "KEY_0"},
		{"1", nil, "KEY_1"},
		{"2", nil, "KEY_2"},
		{"3", nil, "KEY_3"},
		{"4", nil, "KEY_4"},
		{"5", nil, "KEY_5"},
		{"6", nil, "KEY_6"},
		{"7", nil, "KEY_7"},
		{"8", nil, "KEY_8"},
		{"9", nil, "KEY_9"},
		{"+", nil, "KEY_VOLUP"},
		{"-", nil, "KEY_VOLDOWN"},
		{"m", nil, "KEY_MUTE"},
		{"h", nil, "KEY_LEFT"},
		{"l", nil, "KEY_RIGHT"},
		{"j", nil, "KEY_DOWN"},
		{"k", nil, "KEY_UP"},
		{"i", nil, "KEY_HOME"},
		{" ", nil, "KEY_PAUSE"},
		{"p", nil, "KEY_PLAY"},
		{"P", []string{"Shift"}, "KEY_POWER"},
		{"Backspace", nil, "KEY_RETURN"},
		{"Escape", nil, "KEY_RETURN"},
		{"Return", nil, "KEY_ENTER"},

		{"O", []string{"Shift"}, "WOL"},
		{"q", nil, "EXIT"},

		{"H", []string{"Shift"}, "LEFT"},
		{"J", []string{"Shift"}, "DOWN"},
		{"K", []string{"Shift"}, "UP"},
		{"L", []string{"Shift"}, "RIGHT"},
	}

	defaultKeymap = map[int32]string{
		'0':  "KEY_0",
		'1':  "KEY_1",
		'2':  "KEY_2",
		'3':  "KEY_3",
		'4':  "KEY_4",
		'5':  "KEY_5",
		'6':  "KEY_6",
		'7':  "KEY_7",
		'8':  "KEY_8",
		'9':  "KEY_9",
		'+':  "KEY_VOLUP",
		'-':  "KEY_VOLDOWN",
		'm':  "KEY_MUTE",
		'h':  "KEY_LEFT",
		'l':  "KEY_RIGHT",
		'j':  "KEY_DOWN",
		'k':  "KEY_UP",
		'i':  "KEY_HOME",
		' ':  "KEY_PAUSE",
		'p':  "KEY_PLAY",
		'P':  "KEY_POWER",
		0x1B: "KEY_RETURN",
		0x7F: "KEY_RETURN",
		0x0A: "KEY_ENTER",

		'O': "WOL",
		'q': "EXIT",

		'H': "LEFT",
		'J': "DOWN",
		'K': "UP",
		'L': "RIGHT",
	}
)

func readFile(filename string) (b []byte, e error) {
	var (
		fh *os.File
	)

	if fh, e = os.Open(televisionConfigFile); e != nil {
		return
	}
	defer fh.Close()

	b, e = ioutil.ReadAll(fh)

	return
}

func readTelevisionConfig() (config televisions, e error) {
	var (
		b []byte
	)

	if b, e = readFile(televisionConfigFile); e != nil {
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
	if e != nil {
		return
	}

	flags := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if fh, e = os.OpenFile(televisionConfigFile, flags, 0770); e != nil {
		return
	}
	defer fh.Close()

	if b, e = json.MarshalIndent(tvs, "", "  "); e != nil {
		return
	}

	_, e = fh.Write(b)

	return
}
