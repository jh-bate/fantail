package fantail

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"

	"github.com/jh-bate/fantail/data"
	"github.com/jh-bate/fantail/data/notes"
	"github.com/jh-bate/fantail/data/smbgs"
	"github.com/jh-bate/fantail/events"

	"github.com/jh-bate/fantail/users"
)

type Api struct {
	dataStore *data.Store
	eStore    *events.Store
	userStore *users.Store
	Logger    *log.Logger
	Metrics   *log.Logger
	*config
}

type config struct {
	DataStorePath    string `json:"dataStorePath"`
	UserStorePath    string `json:"userStorePath"`
	MetricsStorePath string `json:"metricsStorePath"`
	Secret           string `json:"signingSecret"`
}

func loadConfig() *config {

	_, filename, _, _ := runtime.Caller(1)
	configFile, err := ioutil.ReadFile(path.Join(path.Dir(filename), "apiConfig.json"))

	if err != nil {
		log.Panic("could not load config ", err.Error())
	}
	var apiConf config
	err = json.Unmarshal(configFile, &apiConf)
	if err != nil {
		log.Panic("could not load config")
	}
	return &apiConf
}

func InitApi() *Api {
	usedConfig := loadConfig()

	f, err := os.OpenFile(usedConfig.MetricsStorePath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	return &Api{
		dataStore: data.NewStore(usedConfig.DataStorePath),
		eStore:    events.NewStore(usedConfig.DataStorePath),
		userStore: users.NewStore(usedConfig.UserStorePath),
		Logger:    log.New(os.Stdout, "fantail/api:", log.Lshortfile),
		Metrics:   log.New(f, "fantail/metrics:", log.Ltime|log.Ldate),
		config:    usedConfig,
	}
}

func (a *Api) Login(usr *users.User) (string, error) {
	token := usr.Login(a.Secret)
	if token == "" {
		return token, errors.New("issue trying to login")
	}
	return token, nil
}

func (a *Api) SignupUser(in io.Reader) (*users.User, error) {

	raw := users.DecodeRaw(in)

	a.Logger.Printf("lets check %#v", raw)
	if raw.Valid() {
		savedUser := raw.NewUser()
		a.Logger.Printf("dup check save! %#v", savedUser.Email)
		exists, _ := a.userStore.GetUserByEmail(savedUser.Email)
		if exists != nil {

		}
		a.Logger.Printf("lets save! %#v", savedUser)
		err := a.userStore.AddUser(savedUser)
		if err != nil {
			a.Logger.Println(err.Error())
			return savedUser, ErrInternalServer.Error
		}
		return savedUser, err
	}
	a.Logger.Println(ErrInvalidSignup.Error)
	return nil, ErrInvalidSignup.Error
}

func (a *Api) GetUser(id string) (*users.User, error) {
	if id == "" {
		a.Logger.Println(ErrNoUserId.Error)
		return nil, ErrNoUserId.Error
	}
	foundUser, err := a.userStore.GetUser(id)
	if err != nil {
		a.Logger.Println(err.Error())
		return nil, ErrInternalServer.Error
	}
	return foundUser, nil
}

func (a *Api) AuthenticateUserSession(sessionToken string) (*users.User, error) {

	a.Logger.Println("token ", sessionToken)
	valid, data := users.SessionValid(sessionToken, a.Secret)
	a.Logger.Printf("data %#v", data)

	if valid && data != nil {
		sessionUser, err := a.GetUser(data.UserId)
		if err != nil {
			return nil, errors.New("could not find session user")
		}
		return sessionUser, nil
	}
	return nil, errors.New("invalid or expired session")
}

func (a *Api) RefreshUserSession(sessionToken string) string {
	sessionUser, err := a.AuthenticateUserSession(sessionToken)
	if err != nil {
		a.Logger.Println(err.Error())
		return ""
	}
	return sessionUser.SessionRefresh(sessionToken, a.Secret)
}

func (a *Api) GetUserByEmail(email string) (*users.User, error) {
	if email == "" {
		a.Logger.Println(ErrNoUserId.Error)
		return nil, ErrNoUserId.Error
	}
	foundUser, err := a.userStore.GetUserByEmail(email)
	if err != nil {
		a.Logger.Println(err.Error())
	}
	return foundUser, err
}

func (a *Api) SaveEvents(in io.Reader, out io.Writer, userid string) error {
	if userid == "" {
		a.Logger.Println(ErrNoUserId.Error)
		return ErrNoUserId.Error
	}

	var dbBuffer bytes.Buffer

	events.StreamNew(in, out, &dbBuffer)

	if err := a.eStore.AddEvents(userid, dbBuffer.Bytes()); err != nil {
		a.Logger.Println(err.Error())
		return ErrInternalServer.Error
	}
	return nil
}

func (a *Api) GetEvents(out io.Writer, userid string) error {
	if userid == "" {
		a.Logger.Println(ErrNoUserId.Error)
		return ErrNoUserId.Error
	}

	eventsData, err := a.eStore.GetEvents(userid)
	if err != nil {
		a.Logger.Println(err.Error())
		return ErrInternalServer.Error
	}

	out.Write(eventsData)
	return nil
}

func (a *Api) SaveSmbgs(in io.Reader, out io.Writer, userid string) error {
	if userid == "" {
		a.Logger.Println(ErrNoUserId.Error)
		return ErrNoUserId.Error
	}

	var dbBuffer bytes.Buffer

	smbgs.StreamNew(in, "", "", out, &dbBuffer)

	if err := a.dataStore.AddSmbgs(userid, dbBuffer.Bytes()); err != nil {
		a.Logger.Println(err.Error())
		return ErrInternalServer.Error
	}
	return nil
}

func (a *Api) GetSmbgs(out io.Writer, userid string) error {
	if userid == "" {
		a.Logger.Println(ErrNoUserId.Error)
		return ErrNoUserId.Error
	}

	smbgsData, err := a.dataStore.GetSmbgs(userid)
	if err != nil {
		a.Logger.Println(err.Error())
		return ErrInternalServer.Error
	}

	out.Write(smbgsData)
	return nil
}

func (a *Api) SaveNotes(in io.Reader, out io.Writer, userid string) error {
	if userid == "" {
		a.Logger.Println(ErrNoUserId.Error)
		return ErrNoUserId.Error
	}

	var dbBuffer bytes.Buffer

	notes.StreamNew(in, "", "", out, &dbBuffer)

	if err := a.dataStore.AddNotes(userid, dbBuffer.Bytes()); err != nil {
		a.Logger.Println(err.Error())
		return ErrInternalServer.Error
	}
	return nil
}

func (a *Api) GetNotes(out io.Writer, userid string) error {
	if userid == "" {
		a.Logger.Println(ErrNoUserId.Error)
		return ErrNoUserId.Error
	}

	notesData, err := a.dataStore.GetNotes(userid)
	if err != nil {
		a.Logger.Println(err.Error())
		return ErrInternalServer.Error
	}

	out.Write(notesData)
	return nil
}
