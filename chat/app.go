// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/base64"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/gorilla/mux"

	"github.com/jh-bate/fantail"
	"github.com/jh-bate/fantail/users"
)

type Fantail struct {
	logger *log.Logger
	api    *fantail.Api
}

var f = &Fantail{
	api:    fantail.InitApi(),
	logger: log.New(os.Stdout, "faintail/chat", log.Lshortfile),
}

var addr = flag.String("addr", ":8080", "http service address")
var homeTempl = template.Must(template.ParseFiles("./static/index.html"))

var homeHandler = http.HandlerFunc(serveHome)

func serveHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	homeTempl.Execute(w, r.Host)
}

// serverWs handles websocket requests from the peer.
func serveWs(w http.ResponseWriter, r *http.Request) {

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	c := &connection{send: make(chan []byte, 256), ws: ws, fantail: f}
	h.register <- c
	go c.writePump()
	c.readPump()
}

func loginUser(w http.ResponseWriter, r *http.Request) {

	auth := strings.SplitN(r.Header["Authorization"][0], " ", 2)

	if len(auth) != 2 || auth[0] != "Basic" {
		f.logger.Println("Authorization invalid")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload, _ := base64.StdEncoding.DecodeString(auth[1])
	pair := strings.SplitN(string(payload), ":", 2)

	if len(pair) != 2 {
		f.logger.Println("Authorization invalid")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	usr, err := f.api.GetUserByEmail(pair[0])
	if err != nil {
		f.logger.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}

	if usr != nil && usr.Validate(pair[1]) {
		sessionToken, err := f.api.Login(usr)
		if err != nil {
			f.logger.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set(users.FANTAIL_SESSION_TOKEN, sessionToken)
		return
	}

	w.WriteHeader(http.StatusForbidden)
	return
}

func main() {
	flag.Parse()
	go h.run()

	r := mux.NewRouter()
	r.HandleFunc("/", homeHandler)

	r.HandleFunc("/login", loginUser).Methods("POST")
	r.HandleFunc("/ws/fantail", serveWs).Methods("GET")
	http.Handle("/", r)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
