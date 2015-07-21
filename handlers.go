package fantail

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/daaku/go.httpgzip"

	"github.com/jh-bate/fantail/users"
)

//curl -u jamie@tidepool.org:admin -i -H 'Content-Type: application/json' -d '' http://localhost:8090/login
func (api *Api) LoginHandler(w http.ResponseWriter, r *http.Request) {
	auth := strings.SplitN(r.Header["Authorization"][0], " ", 2)

	if len(auth) != 2 || auth[0] != "Basic" {
		api.Logger.Println("Authorization invalid")
		api.WriteError(w, ErrAuthHeader.SetId())
		return
	}

	payload, _ := base64.StdEncoding.DecodeString(auth[1])
	pair := strings.SplitN(string(payload), ":", 2)

	if len(pair) != 2 {
		api.Logger.Println("Authorization invalid")
		api.WriteError(w, ErrAuthHeader.SetId())
		return
	}

	usr, err := api.GetUserByEmail(pair[0])

	if err != nil {
		api.Logger.Println(err.Error())
		api.WriteError(w, ErrInternalServer.SetId())
		return
	}

	if usr != nil && usr.Validate(pair[1]) {
		sessionToken, err := api.Login(usr)
		if err != nil {
			api.Logger.Println(err.Error())
			api.WriteError(w, ErrInternalServer.SetId())
			return
		}
		api.Logger.Println("we have a token")
		w.Header().Set(users.FANTAIL_SESSION_TOKEN, sessionToken)
		return
	}
	w.WriteHeader(http.StatusForbidden)
	return
}

//curl -d '{"email": "jamie@tidepool.org","name":"Jamie Bate", "password": "admin"}' -H "Content-Type:application/json" http://localhost:8090/signup
func (api *Api) SignupHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	savedUsr, err := api.SignupUser(r.Body)
	if err != nil {
		api.Logger.Println(err.Error())
		api.WriteError(w, ErrInvalidRequest.SetId())
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(savedUsr.ToPublishedUser())
	return
}

func (api *Api) RefreshSessionHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if newToken := api.RefreshUserSession(r.Header.Get(users.FANTAIL_SESSION_TOKEN)); newToken != "" {
			w.Header().Set(users.FANTAIL_SESSION_TOKEN, newToken)
			next.ServeHTTP(w, r)
		}
		w.WriteHeader(http.StatusUnauthorized)
		return
	})
}

func (api *Api) AuthHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if usr, err := api.AuthenticateUserSession(r.Header.Get(users.FANTAIL_SESSION_TOKEN)); err != nil {
			api.Logger.Println(err.Error())
			api.WriteError(w, ErrInternalServer.SetId())
			return
		} else if usr != nil {
			api.Logger.Println("user authenticated")
			next.ServeHTTP(w, r)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		return
	})
}

func (api *Api) LoggingHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()
		log.Printf("[%s] %q %v\n", r.Method, r.URL.String(), t2.Sub(t1))
	}
	return http.HandlerFunc(fn)
}

func (api *Api) WriteError(w http.ResponseWriter, err *DetailedError) {
	api.Logger.Printf("writing error %v", err)
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(err.Status)
	json.NewEncoder(w).Encode(Errors{[]*DetailedError{err}})
}

func (api *Api) WriteData(w http.ResponseWriter, jsonData []byte) {
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.Write(jsonData)
}

func (api *Api) GzipHandler(h http.Handler) http.Handler {
	return httpgzip.NewHandler(h)
}
