// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/jh-bate/fantail"
	"github.com/jh-bate/fantail/users"
)

type fantailApi struct {
	logger *log.Logger
	api    *fantail.Api
}

var fApi = &fantailApi{
	api:    fantail.InitApi(),
	logger: log.New(os.Stdout, "faintail/chat", log.Lshortfile),
}

var addr = flag.String("addr", ":8080", "http service address")
var homeTempl = template.Must(template.ParseFiles("home.html"))
var homeHandler = http.HandlerFunc(serveHome)

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	homeTempl.Execute(w, r.Host)
}

// serverWs handles websocket requests from the peer.
func serveWs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	c := &connection{send: make(chan []byte, 256), ws: ws}
	h.register <- c
	go c.writePump(fApi)
	c.readPump()
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Executing authMiddleware")
		if usr, err := fApi.api.AuthenticateUserSession(r.Header.Get(users.FANTAIL_SESSION_TOKEN)); err != nil {
			fApi.logger.Println(err.Error())
		} else if usr != nil {
			fApi.logger.Println("user authenticated")
			next.ServeHTTP(w, r)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		return
	})
}

func main() {
	flag.Parse()
	go h.run()

	http.Handle("/", homeHandler)
	http.HandleFunc("/ws", serveWs)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
