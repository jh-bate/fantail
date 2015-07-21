package main

import (
	"bytes"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jh-bate/fantail"
	"github.com/justinas/alice"
)

var fantailApi = fantail.InitApi()

func main() {

	commonHandlers := alice.New(fantailApi.LoggingHandler, fantailApi.AuthHandler, fantailApi.GzipHandler)

	r := mux.NewRouter()

	r.HandleFunc("/login", fantailApi.LoginHandler).Methods("POST")
	r.HandleFunc("/signup", fantailApi.SignupHandler).Methods("POST")

	r.Handle("/data/{userid}", commonHandlers.ThenFunc(getEvents)).Methods("GET")
	r.Handle("/data/{userid}", commonHandlers.ThenFunc(postEvents)).Methods("POST")

	//r.Handle("/data/{userid}/notes", commonHandlers.ThenFunc(getNotes)).Methods("GET")
	//r.Handle("/data/{userid}/smbgs", commonHandlers.ThenFunc(getSmbgs)).Methods("GET")
	//r.Handle("/data/{userid}/notes", commonHandlers.ThenFunc(postNotes)).Methods("POST")
	//r.Handle("/data/{userid}/smbgs", commonHandlers.ThenFunc(postSmbgs)).Methods("POST")

	http.Handle("/", r)
	fantailApi.Logger.Println(http.ListenAndServe(":8090", nil))
}

func notImplemented(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	return
}

func getEvents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userid := vars["userid"]

	var eventsBuffer bytes.Buffer

	err := fantailApi.GetEvents(&eventsBuffer, userid)

	if err != nil {
		fantailApi.Logger.Println(err.Error())
		fantailApi.WriteError(w, fantail.ErrInternalServer.SetId())
		return
	}
	w.Write(eventsBuffer.Bytes())
	return
}

func postEvents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userid := vars["userid"]

	var confirmationBuffer bytes.Buffer
	err := fantailApi.SaveEvents(r.Body, &confirmationBuffer, userid)

	if err != nil {
		fantailApi.Logger.Println(err.Error())
		fantailApi.WriteError(w, fantail.ErrInternalServer.SetId())
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(confirmationBuffer.Bytes())
	return
}

/*func getSmbgs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userid := vars["userid"]

	var smbgsBuffer bytes.Buffer

	err := fantailApi.GetSmbgs(&smbgsBuffer, userid)

	if err != nil {
		fantailApi.Logger.Println(err.Error())
		fantailApi.WriteError(w, fantail.ErrInternalServer.SetId())
		return
	}
	w.Write(smbgsBuffer.Bytes())
	return
}

func postSmbgs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userid := vars["userid"]

	var confirmationBuffer bytes.Buffer
	err := fantailApi.SaveSmbgs(r.Body, &confirmationBuffer, userid)

	if err != nil {
		fantailApi.Logger.Println(err.Error())
		fantailApi.WriteError(w, fantail.ErrInternalServer.SetId())
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(confirmationBuffer.Bytes())
	return
}

func getNotes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userid := vars["userid"]

	var notesBuffer bytes.Buffer

	err := fantailApi.GetNotes(&notesBuffer, userid)

	if err != nil {
		fantailApi.Logger.Println(err.Error())
		fantailApi.WriteError(w, fantail.ErrInternalServer.SetId())
		return
	}
	w.Write(notesBuffer.Bytes())
	return
}

func postNotes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userid := vars["userid"]

	var confirmationBuffer bytes.Buffer
	err := fantailApi.SaveNotes(r.Body, &confirmationBuffer, userid)

	if err != nil {
		fantailApi.Logger.Println(err.Error())
		fantailApi.WriteError(w, fantail.ErrInternalServer.SetId())
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(confirmationBuffer.Bytes())
	return
}*/
