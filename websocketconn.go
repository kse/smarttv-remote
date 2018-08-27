package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// Continuously read from websocket until error is returned.
func loopreader(conn *websocket.Conn) {
	for {
		_, _, e := conn.ReadMessage()
		if e != nil {
			log.Println(e.Error())
			return
		}

		// Ignore returned message, for now.
		//fmt.Printf("Response: %s\n", r)
	}
}

func connectToTV() (conn *websocket.Conn, e error) {
	dialer := &websocket.Dialer{}
	dialer.HandshakeTimeout = time.Millisecond * 500
	conn, response, e := dialer.Dial(tvAddr, nil)
	if e != nil {
		return
	}

	_, e = ioutil.ReadAll(response.Body)
	response.Body.Close()
	if e != nil {
		return
	}

	return
}

func messenger(ch <-chan string) {
	var (
		conn *websocket.Conn
		e    error
		name = base64.StdEncoding.EncodeToString([]byte("samsungctl"))
	)

	tvAddr = fmt.Sprintf(tvAddr, ip, name)

	conn, e = connectToTV()
	if e == nil {
		go loopreader(conn)
	}

	for command := range ch {
		if conn == nil {
			// Skip the command if we can't connect to the TV.
			conn, e = connectToTV()
			if e != nil {
				conn = nil
				log.Println(e.Error())
				continue
			}

			go loopreader(conn)
		}

		payload := payload{
			Method: "ms.remote.control",
			Params: params{
				Cmd:          "Click",
				DataOfCmd:    command,
				Option:       "false",
				TypeOfRemote: "SendRemoteKey",
			},
		}

		js, e := json.Marshal(payload)
		if e != nil {
			log.Println(e.Error())
			conn = nil
			continue
		}

		conn.SetWriteDeadline(time.Now().Add(time.Millisecond * 500))
		e = conn.WriteMessage(websocket.TextMessage, js)
		if e != nil {
			log.Println(e.Error())
			conn = nil
			continue
		}
	}
}
