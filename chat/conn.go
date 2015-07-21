// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jh-bate/fantail"
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

	api *fantail.Api

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

	dataUsr, err := c.api.AuthenticateUserSession(data["user"].(string))

	if dataUsr != nil {
		delete(data, "user")
		data["creatorId"] = dataUsr.Id
		if jsonData, err := json.Marshal(data); err == nil {

			eventStr := string(jsonData[:])

			c.api.SaveEvents(strings.NewReader(eventStr), os.Stdout, dataUsr.Id)
			if strings.Contains(strings.ToLower(data["type"].(string)), "note") {
				note := data["data"].(map[string]interface{})
				return note["text"].(string), nil
			} else if strings.Contains(strings.ToLower(data["type"].(string)), "smbg") {
				smbg := data["data"].(map[string]interface{})["value"].(float64)

				if smbg > 10 {
					return fmt.Sprintf("%.1f", smbg), []string{"good work on taking a BG", "was that expected or un-expected?", "any notes you would like to add?"}
				} else if smbg < 4 {
					return fmt.Sprintf("%.1f", smbg), []string{"good work on taking a BG", "time to eat?", "remember to retest after 15 mins of treating a low", "and maybe add a note later so we can workout what might have gone wrong"}
				}

				return fmt.Sprintf("%.1f", smbg), []string{"awesome work - anything that is note worthy?"}
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
