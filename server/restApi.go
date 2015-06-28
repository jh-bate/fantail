package main

import (
	"bytes"
	"encoding/base64"
	"log"
	"net/http"
	"strings"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/jh-bate/d-data-cli/client"
	"github.com/jh-bate/d-data-cli/user"
)

var dataApi *client.Api
var userStore *user.Store

const session_token = "x-dhub-token"

func main() {

	dataApi = client.InitApi(client.NewStore())
	userStore = &user.Store{Users: map[string]*user.User{}}

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
	//if user.SessionValid(r.Header.Get(session_token), userStore) {
	handler(w, r)
	//	return
	//}
	//w.WriteHeader(http.StatusUnauthorized)
}

func refresh(w rest.ResponseWriter, r *rest.Request) {

	if updated := user.SessionRefresh(r.Header.Get(session_token), userStore); updated != "" {
		w.Header().Set(session_token, updated)
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

	if usr := userStore.GetUser(pair[0]); usr != nil {
		if usr.Validate(pair[1]) {
			log.Println("logging in ...")
			token := usr.Login(userStore)
			log.Println("logged in ", token)
			w.Header().Set(session_token, token)
			return
		}
		w.WriteHeader(http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusForbidden)
	return
}

//curl -d '{"email": "jamie@tidepool.org","name":"Jamie Bate", "password": "admin"}' -H "Content-Type:application/json" http://localhost:8090/signup
func signup(w rest.ResponseWriter, r *rest.Request) {

	raw := user.DecodeRaw(r.Body)

	if raw.Email == "" || raw.Name == "" || raw.Pass == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newUsr := user.NewUser(raw.Name, raw.Email, raw.Pass)

	if newUsr.Signup(userStore) {
		w.WriteHeader(http.StatusCreated)
		w.WriteJson(newUsr)
		return
	}

	w.WriteHeader(http.StatusConflict)
	w.WriteJson(newUsr)
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
