package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

var (
	sdlKeymap map[sdl.Keycode][]*parsedKeyInput
)

type parsedKeyInput struct {
	KeyCode sdl.Keycode
	Mod     uint16
	Target  string
}

func parseKeyInput(k keyInput) (o *parsedKeyInput, e error) {
	o = &parsedKeyInput{}
	o.KeyCode = sdl.GetKeyFromName(k.key)
	o.Target = k.event

	for _, m := range k.mod {
		switch m {
		case "Shift":
			o.Mod |= uint16(sdl.KMOD_SHIFT)
		}
	}

	return
}

func generateParsedKeymap() {
	sdlKeymap = make(map[sdl.Keycode][]*parsedKeyInput)
	for _, k := range sdlDefaultKeymap {
		p, e := parseKeyInput(k)
		if e != nil {
			continue
		}

		_, ok := sdlKeymap[p.KeyCode]

		if !ok {
			sdlKeymap[p.KeyCode] = make([]*parsedKeyInput, 0)
		}

		sdlKeymap[p.KeyCode] = append(sdlKeymap[p.KeyCode], p)
	}
}
