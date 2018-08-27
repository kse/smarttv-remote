package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
)

// Wake-On-LAN
// Packet is UDP to port 0,
// FF FF FF FF FF FF
// followed by target MAC address 16 times as content.
// Broadcast to network.
func wol(mac string) {
	addr := &net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: 0,
	}

	conn, e := net.DialUDP("udp", nil, addr)
	if e != nil {
		fmt.Println(e.Error())
		return
	}

	buffer := &bytes.Buffer{}

	for i := 0; i < 6; i++ {
		buffer.WriteByte(byte(255))
	}

	hexString := strings.Replace(mac, ":", "", -1)

	macBytes, e := hex.DecodeString(hexString)
	if e != nil {
		fmt.Println(e.Error())
		return
	}

	for i := 0; i < 16; i++ {
		_, e := buffer.Write(macBytes)
		if e != nil {
			fmt.Println(e.Error())
			return
		}
	}

	conn.Write(buffer.Bytes())
}
