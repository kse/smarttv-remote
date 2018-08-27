package main

import (
	"fmt"
	"log"
	"os"
)

var (
	logger = log.Logger{}
	mac    string
	ip     string
	tvAddr = "ws://%s:8001/api/v2/channels/samsung.remote.control?name=%s"

	tvsConfig televisions
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

func main() {
	var (
		e   error
		ch  = make(chan string)
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

		switch buf[0] {
		case 'q':
			return
		case 'a':
			ch <- "KEY_ASPECT"
		case '0':
			ch <- "KEY_0"
		case '1':
			ch <- "KEY_1"
		case '2':
			ch <- "KEY_2"
		case '3':
			ch <- "KEY_3"
		case '4':
			ch <- "KEY_4"
		case '5':
			ch <- "KEY_5"
		case '6':
			ch <- "KEY_6"
		case '7':
			ch <- "KEY_7"
		case '8':
			ch <- "KEY_8"
		case '9':
			ch <- "KEY_9"
		case '+':
			ch <- "KEY_VOLUP"
		case '-':
			ch <- "KEY_VOLDOWN"
		case 'm':
			ch <- "KEY_MUTE"
		case 'h':
			ch <- "KEY_LEFT"
		case 'l':
			ch <- "KEY_RIGHT"
		case 'j':
			ch <- "KEY_DOWN"
		case 'k':
			ch <- "KEY_UP"
		case 'i':
			ch <- "KEY_HOME"
		case ' ':
			ch <- "KEY_PAUSE"
		case 'p':
			ch <- "KEY_PLAY"
		case 'P':
			ch <- "KEY_POWER"
		case 'O':
			wol(mac)
		case 0x1B:
			fallthrough
		case 0x7F:
			ch <- "KEY_RETURN"
		case 0x0A:
			ch <- "KEY_ENTER"
		default:
			//fmt.Println(buf[0])
		}
	}
}
