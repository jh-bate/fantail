package main

import (
	"bytes"
	"encoding/base64"
	"log"
	"net/http"
	"strings"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/jh-bate/fantail/client"
)

var dataApi *client.Api

//var userStore *user.Store

const session_token = "x-dhub-token"

func main() {

	dataApi = client.InitApi(client.NewStore())
	//userStore = &user.Store{Users: map[string]*user.User{}}

	api := rest.NewApi()

	statusMw := &rest.StatusMiddleware{}
	api.Use(statusMw)

	api.Use(&rest.GzipMiddleware{})
	api.Use(&rest.AccessLogJsonMiddleware{})

	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		//used for service status reporting
		rest.Get("/data/.status", func(w rest.ResponseWriter, r *rest.Request) {
			w.WriteJson(statusMw.GetStatus())
		}),

		rest.Post("/login", login),
		rest.Post("/refresh", refresh),
		rest.Post("/signup", signup),

		rest.Get("/data/:userid/smbgs", func(w rest.ResponseWriter, r *rest.Request) {
			checkAuth(w, r, getSmbgs)
		}),
		rest.Post("/data/:userid/smbgs", func(w rest.ResponseWriter, r *rest.Request) {
			checkAuth(w, r, postSmbgs)
		}),
		rest.Put("/data/:userid/smbgs", notImplemented),

		rest.Get("/data/:userid/notes", notImplemented),
		rest.Post("/data/:userid/notes", notImplemented),
		rest.Put("/data/:userid/notes", notImplemented),

		rest.Get("/data/:userid/basals", notImplemented),
		rest.Post("/data/:userid/basals", notImplemented),
		rest.Put("/data/:userid/basals", notImplemented),

		rest.Get("/data/:userid/boluses", notImplemented),
		rest.Post("/data/:userid/boluses", notImplemented),
		rest.Put("/data/:userid/boluses", notImplemented),

		rest.Get("/data/:userid/cbgs", notImplemented),
		rest.Post("/data/:userid/cbgs", notImplemented),
		rest.Put("/data/:userid/cbgs", notImplemented),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":8090", api.MakeHandler()))
}

func notImplemented(w rest.ResponseWriter, r *rest.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	return
}

func checkAuth(w rest.ResponseWriter, r *rest.Request, handler rest.HandlerFunc) {

	usr, _ := dataApi.AuthenticateUserSession(r.Header.Get(session_token))
	log.Printf("authenticated user? %#v", usr)
	/*
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}*/
	handler(w, r)
}

func refresh(w rest.ResponseWriter, r *rest.Request) {
	if newToken := dataApi.RefreshUserSession(r.Header.Get(session_token)); newToken != "" {
		w.Header().Set(session_token, newToken)
		return
	}
	w.WriteHeader(http.StatusUnauthorized)
	return
}

//curl -u jamie@tidepool.org:admin -i -H 'Content-Type: application/json' -d '' http://localhost:8090/login
func login(w rest.ResponseWriter, r *rest.Request) {

	auth := strings.SplitN(r.Header["Authorization"][0], " ", 2)

	if len(auth) != 2 || auth[0] != "Basic" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload, _ := base64.StdEncoding.DecodeString(auth[1])
	pair := strings.SplitN(string(payload), ":", 2)

	if len(pair) != 2 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	usr, err := dataApi.GetUserByEmail(pair[0])

	if err != nil {
		log.Println("login err:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}

	if usr != nil && usr.Validate(pair[1]) {
		sessionToken := usr.Login()
		log.Println("logged in ", sessionToken)
		w.Header().Set(session_token, sessionToken)
		return
	}

	w.WriteHeader(http.StatusForbidden)
	return
}

//curl -d '{"email": "jamie@tidepool.org","name":"Jamie Bate", "password": "admin"}' -H "Content-Type:application/json" http://localhost:8090/signup
func signup(w rest.ResponseWriter, r *rest.Request) {

	savedUsr, err := dataApi.SaveUser(r.Body)
	if err != nil {
		log.Println("signup err: ", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.WriteJson(savedUsr)
	return
}

func getSmbgs(w rest.ResponseWriter, r *rest.Request) {
	userid := r.PathParam("userid")

	var smbgsBuffer bytes.Buffer

	err := dataApi.GetSmbgs2(&smbgsBuffer, userid)

	//log.Println("getSmbgs ", string(smbgsBuffer.Bytes()[:]))

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonErr, _ := w.EncodeJson(err)
		w.WriteJson(jsonErr)
		return
	}
	w.(http.ResponseWriter).Write(smbgsBuffer.Bytes())
	return
}

func postSmbgs(w rest.ResponseWriter, r *rest.Request) {

	userid := r.PathParam("userid")

	var confirmationBuffer bytes.Buffer
	err := dataApi.SaveSmbgs2(r.Body, &confirmationBuffer, userid)

	//log.Println("postSmbgs confirmation ", string(confirmationBuffer.Bytes()[:]))

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonErr, _ := w.EncodeJson(err)
		w.WriteJson(jsonErr)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.(http.ResponseWriter).Write(confirmationBuffer.Bytes())
	return
}
