package main

import (
	"bytes"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/jh-bate/fantail"

	"github.com/jh-bate/fantail/users"
)

type fantailApi struct {
	logger *log.Logger
	api    *fantail.Api
}

var fApi = &fantailApi{
	api:    fantail.InitApi(),
	logger: log.New(os.Stdout, "faintail/api", log.Lshortfile),
}

func main() {

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
	log.Println(http.ListenAndServe(":8090", api.MakeHandler()))
}

func notImplemented(w rest.ResponseWriter, r *rest.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	return
}

func checkAuth(w rest.ResponseWriter, r *rest.Request, handler rest.HandlerFunc) {

	if usr, err := fApi.api.AuthenticateUserSession(r.Header.Get(users.FANTAIL_SESSION_TOKEN)); err != nil {
		fApi.logger.Println(err.Error())
	} else if usr != nil {
		fApi.logger.Println("user authenticated")
		handler(w, r)
		return
	}
	w.WriteHeader(http.StatusUnauthorized)
	return
}

func refresh(w rest.ResponseWriter, r *rest.Request) {
	if newToken := fApi.api.RefreshUserSession(r.Header.Get(users.FANTAIL_SESSION_TOKEN)); newToken != "" {
		w.Header().Set(users.FANTAIL_SESSION_TOKEN, newToken)
		return
	}
	w.WriteHeader(http.StatusUnauthorized)
	return
}

//curl -u jamie@tidepool.org:admin -i -H 'Content-Type: application/json' -d '' http://localhost:8090/login
func login(w rest.ResponseWriter, r *rest.Request) {

	auth := strings.SplitN(r.Header["Authorization"][0], " ", 2)

	if len(auth) != 2 || auth[0] != "Basic" {
		fApi.logger.Println("Authorization invalid")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload, _ := base64.StdEncoding.DecodeString(auth[1])
	pair := strings.SplitN(string(payload), ":", 2)

	if len(pair) != 2 {
		fApi.logger.Println("Authorization invalid")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	usr, err := fApi.api.GetUserByEmail(pair[0])

	if err != nil {
		fApi.logger.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}

	if usr != nil && usr.Validate(pair[1]) {
		sessionToken, err := fApi.api.Login(usr)
		if err != nil {
			fApi.logger.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fApi.logger.Println("we have a token")
		w.Header().Set(users.FANTAIL_SESSION_TOKEN, sessionToken)
		return
	}

	w.WriteHeader(http.StatusForbidden)
	return
}

//curl -d '{"email": "jamie@tidepool.org","name":"Jamie Bate", "password": "admin"}' -H "Content-Type:application/json" http://localhost:8090/signup
func signup(w rest.ResponseWriter, r *rest.Request) {

	savedUsr, err := fApi.api.SignupUser(r.Body)
	if err != nil {
		fApi.logger.Println(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.WriteJson(savedUsr.ToPublishedUser())
	return
}

func getSmbgs(w rest.ResponseWriter, r *rest.Request) {
	userid := r.PathParam("userid")

	var smbgsBuffer bytes.Buffer

	err := fApi.api.GetSmbgs(&smbgsBuffer, userid)

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
	err := fApi.api.SaveSmbgs(r.Body, &confirmationBuffer, userid)

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
