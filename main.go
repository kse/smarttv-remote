package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"encoding/json"

	"github.com/veandco/go-sdl2/sdl"
)

var (
	logger = log.Logger{}
	mac    string
	ip     string
	tvAddr = "ws://%s:8001/api/v2/channels/samsung.remote.control?name=%s"

	tvsConfig televisions

	mouseStep  = 20
	mouseScale = 3
)

type params struct {
	Cmd          string
	DataOfCmd    string
	Option       string
	TypeOfRemote string
}

type payload struct {
	Method string `json:"method"`
	Params params `json:"params"`
}

type tvInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Device  struct {
		Type      string `json:"type"`
		ModelName string `json:"modelName"`
		IP        string `json:"ip"`
		WifiMAC   string `json:"wifiMac"`
	} `json:"device"`
}

type networkInfo struct {
	DefaultGateway string
	GatewayMac     string
	Interface      string
}

func findTelevision(mac string, tvs televisions) *tvnetInfo {
	for _, tv := range *tvs {
		if tv.Network.GatewayMac == mac {
			return &tv
		}
	}

	return nil
}

func buildKeyPayloadString(key string) (js []byte) {
	payload := payload{
		Method: "ms.remote.control",
		Params: params{
			Cmd:          "Click",
			DataOfCmd:    key,
			Option:       "false",
			TypeOfRemote: "SendRemoteKey",
		},
	}

	js, _ = json.Marshal(payload)

	return
}

func ts() string {
	return strconv.FormatInt(time.Now().UnixNano()/1000/1000, 10)
}

func buildMousePayload(x, y int) (b []byte) {
	cmd := fmt.Sprintf(`{"method":"ms.remote.control","params":{"Cmd":"Move","Position":{"x":%d,"y":%d,"Time":"%s"},"TypeOfRemote":"ProcessMouseDevice"}}`, x, y, ts())
	return []byte(cmd)
}

func initTerminal() {
	var (
		ch = make(chan []byte)
	)

	// Set terminal to raw input, normal output
	prev, e := terminalRawInput(nil)
	if e != nil {
		log.Fatalln(e.Error())
	}

	// Restore terminal function
	defer terminalRawInput(prev)

	// Send messages to the television
	go messenger(ch)

	//fmt.Println("Entering raw mode")
	buf := make([]byte, 1)
	for {
		_, e = os.Stdin.Read(buf)
		if e != nil {
			return
		}

		var input = int32(buf[0])

		cmd, ok := defaultKeymap[input]

		if !ok {
			fmt.Println(input)
			continue
		}

		if strings.HasPrefix(cmd, "KEY_") {
			b := buildKeyPayloadString(cmd)
			ch <- b
		} else {
			switch cmd {
			case "WOL":
				wol(mac)
			case "EXIT":
				return
			case "LEFT":
				b := buildMousePayload(-mouseStep, 0)
				ch <- b
			case "RIGHT":
				b := buildMousePayload(mouseStep, 0)
				ch <- b
			case "UP":
				b := buildMousePayload(0, -mouseStep)
				ch <- b
			case "DOWN":
				b := buildMousePayload(0, mouseStep)
				ch <- b
			}
		}
	}
}

func initSDL() int {
	var (
		ch                        = make(chan []byte)
		winTitle                  = "SmartTV Remote"
		winWidth, winHeight int32 = 400, 600
		window              *sdl.Window
		event               sdl.Event
		running             bool
		err                 error
	)

	// Send messages to the television
	go messenger(ch)

	sdl.Init(sdl.INIT_EVERYTHING)
	defer sdl.Quit()

	window, err = sdl.CreateWindow(winTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		winWidth, winHeight, sdl.WINDOW_SHOWN|sdl.WINDOW_MOUSE_CAPTURE)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		return 1
	}
	defer window.Destroy()

	sdl.SetRelativeMouseMode(true)

	running = true
	for running {
		for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.MouseMotionEvent:
				fmt.Printf("[%d ms] MouseMotion\ttype:%d\tid:%d\tx:%d\ty:%d\txrel:%d\tyrel:%d\n",
					t.Timestamp, t.Type, t.Which, t.X, t.Y, t.XRel, t.YRel)
				b := buildMousePayload(int(t.XRel)*mouseScale, int(t.YRel)*mouseScale)
				ch <- b
			case *sdl.MouseButtonEvent:
				fmt.Printf("[%d ms] MouseButton\ttype:%d\tid:%d\tx:%d\ty:%d\tbutton:%d\tstate:%d\n",
					t.Timestamp, t.Type, t.Which, t.X, t.Y, t.Button, t.State)
				/*
					case *sdl.MouseWheelEvent:
						fmt.Printf("[%d ms] MouseWheel\ttype:%d\tid:%d\tx:%d\ty:%d\n",
							t.Timestamp, t.Type, t.Which, t.X, t.Y)
				*/
			case *sdl.KeyboardEvent:
				fmt.Printf("[%d ms] Keyboard\ttype:%d\tsym:%c\tmodifiers:%d\tstate:%d\trepeat:%d\n",
					t.Timestamp, t.Type, t.Keysym.Sym, t.Keysym.Mod, t.State, t.Repeat)

				cmd, ok := defaultKeymap[int32(t.Keysym.Sym)]

				if !ok {
					fmt.Println(t.Keysym.Sym)
					continue
				}
				if strings.HasPrefix(cmd, "KEY_") {
					b := buildKeyPayloadString(cmd)
					ch <- b
				} else {
					switch cmd {
					case "WOL":
						wol(mac)
					case "EXIT":
						os.Exit(0)
					case "LEFT":
						b := buildMousePayload(-mouseStep, 0)
						ch <- b
					case "RIGHT":
						b := buildMousePayload(mouseStep, 0)
						ch <- b
					case "UP":
						b := buildMousePayload(0, -mouseStep)
						ch <- b
					case "DOWN":
						b := buildMousePayload(0, mouseStep)
						ch <- b
					}
				}
			}
		}

		sdl.Delay(16)
	}

	return 0
}

func main() {
	var (
		e   error
		tvs televisions
		tv  *tvnetInfo
	)

	netinfo, e := findNetworkInterface()
	if e != nil {
		log.Fatalln(e.Error())
	}

	if tvs, e = readTelevisionConfig(); e != nil {
		fmt.Println("No television configuration found")
		tvs = new([]tvnetInfo)
	} else {
		if tv = findTelevision(netinfo.GatewayMac, tvs); tv == nil {
			fmt.Println("TV for network not found in configuration")
		}
	}

	if tv == nil {
		fmt.Println("Searching for television...")
		tvinfo, e := waitForUDPMulticast(netinfo.Interface)
		if e != nil {
			log.Fatalln(e.Error())
		}

		tv = &tvnetInfo{tvinfo, netinfo}
		*tvs = append(*tvs, tvnetInfo{tvinfo, netinfo})

		fmt.Printf("Found %s\n", tv.Tv.Name)

		e = writeTelevisionConfig(tvs)
		if e != nil {
			log.Println("Unable to write configuration")
			log.Println(e.Error())
		}
	}

	fmt.Printf("Connecting to %s (%s)\n", tv.Tv.Name, tv.Tv.Device.IP)

	ip = tv.Tv.Device.IP
	mac = tv.Tv.Device.WifiMAC

	initSDL()
	//initTerminal()
}
