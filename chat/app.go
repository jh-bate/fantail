// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"

	"github.com/jh-bate/fantail"
)

var api = fantail.InitApi()

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
	c := &connection{send: make(chan []byte, 256), ws: ws, api: api}
	h.register <- c
	go c.writePump()
	c.readPump()
}

func main() {
	flag.Parse()
	go h.run()

	commonHandlers := alice.New(api.LoggingHandler)

	r := mux.NewRouter()
	r.HandleFunc("/", homeHandler)

	r.Handle("/login", commonHandlers.ThenFunc(api.LoginHandler)).Methods("POST")
	r.HandleFunc("/ws/fantail", serveWs).Methods("GET")
	http.Handle("/", r)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
