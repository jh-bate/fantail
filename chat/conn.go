// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// connection is an middleman between the websocket connection and the hub.
type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	fantail *Fantail

	// Buffered channel of outbound messages.
	send chan []byte
}

// readPump pumps messages from the websocket connection to the hub.
func (c *connection) readPump() {
	defer func() {
		h.unregister <- c
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		h.broadcast <- message
	}
}

// write writes a message with the given message type and payload.
func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

func (c *connection) saveData(rawData []byte) (display string, feedback []string) {
	var data map[string]interface{}
	json.Unmarshal(rawData, &data)

	dataUsr, err := c.fantail.api.AuthenticateUserSession(data["user"].(string))

	if dataUsr != nil {
		delete(data, "user")
		data["creatorId"] = dataUsr.Id
		if jsonData, err := json.Marshal(data); err == nil {

			eventStr := string(jsonData[:])

			if strings.Contains(strings.ToLower(eventStr), "note") {
				c.fantail.api.SaveNotes(strings.NewReader(eventStr), os.Stdout, dataUsr.Id)
				return data["text"].(string), nil
			} else if strings.Contains(strings.ToLower(eventStr), "smbg") {
				c.fantail.api.SaveSmbgs(strings.NewReader(eventStr), os.Stdout, dataUsr.Id)

				smbgVal, _ := strconv.ParseFloat(data["value"].(string), 64)

				if smbgVal > 10 {
					return data["value"].(string), []string{"thanks ...", "any notes to add?", "any changes in your routine?"}
				} else if smbgVal < 4 {
					return data["value"].(string), []string{"time to eat?", "remember to retest after 15 mins of treating a low", "and maybe add a note later"}
				}

				return data["value"].(string), []string{"nice work!", "is anything thats worth noting?"}
			}
			return "hmmmm something went wrong there!", nil
		}
	}
	return err.Error(), nil
}

// writePump pumps messages from the hub to the websocket connection.
func (c *connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			//save to api
			chatMessage, feedback := c.saveData(message)
			//lets chat!
			if err := c.write(websocket.TextMessage, []byte(chatMessage)); err != nil {
				return
			}
			for i := range feedback {
				time.Sleep(time.Second * 1) //brief pause
				if err := c.write(websocket.TextMessage, []byte(feedback[i])); err != nil {
					return
				}
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
