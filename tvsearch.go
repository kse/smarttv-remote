package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os/exec"
)

func waitForUDPMulticast(ifaceName string) (info tvInfo, e error) {
	iface, e := net.InterfaceByName(ifaceName)
	if e != nil {
		return
	}

	udpaddr := &net.UDPAddr{}
	udpaddr.IP = net.IPv4(239, 255, 255, 250)
	udpaddr.Port = 15600

	conn, e := net.ListenMulticastUDP("udp", iface, udpaddr)
	if e != nil {
		return
	}
	defer conn.Close()

	b := make([]byte, 2048)
	for {
		e = nil // Reset e, so it is nil if we return
		_, raddr, e := conn.ReadFromUDP(b)
		if e != nil {
			log.Fatalln(e.Error())
		}

		infoURL := fmt.Sprintf("http://%s:8001/api/v2/", raddr.IP)
		resp, e := http.Get(infoURL)

		if e != nil {
			log.Println(e.Error())
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			log.Println("Invalid response code", resp.Status)
			continue
		}

		body, e := ioutil.ReadAll(resp.Body)
		if e != nil {
			log.Println(e.Error())
			continue
		}

		if e = json.Unmarshal(body, &info); e != nil {
			log.Println(e.Error())
			continue
		}

		if info.Device.Type == "Samsung SmartTV" {
			break
		}
	}

	return
}

func findNetworkInterface() (ninfo networkInfo, e error) {
	var (
		output []byte
		cmd    *exec.Cmd
		split  [][]byte
	)
	cmd = exec.Command("ip", "route", "show", "default")
	if output, e = cmd.Output(); e != nil {
		return
	}

	split = bytes.Split(output, []byte(" "))
	if len(split) < 5 {
		e = fmt.Errorf("Unable to determine default route from: '%s'", output)
		return
	}

	ninfo.DefaultGateway = string(split[2])
	ninfo.Interface = string(split[4])

	cmd = exec.Command("ip", "neigh", "show", ninfo.DefaultGateway)
	if output, e = cmd.Output(); e != nil {
		return
	}

	split = bytes.Split(output, []byte(" "))
	if len(split) < 6 {
		e = fmt.Errorf("Neigh output too short: %s", output)
		return
	}

	ninfo.GatewayMac = string(split[4])

	return
}
