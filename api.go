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
	"github.com/jh-bate/fantail/data/smbg"

	"github.com/jh-bate/fantail/users"
)

type Api struct {
	dataStore *data.Store
	userStore *users.Store
	logger    *log.Logger
	*config
}

type config struct {
	DataStorePath string `json:"dataStorePath"`
	UserStorePath string `json:"userStorePath"`
	Secret        string `json:"signingSecret"`
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
	return &Api{
		dataStore: data.NewStore(usedConfig.DataStorePath),
		userStore: users.NewStore(usedConfig.UserStorePath),
		logger:    log.New(os.Stdout, "fantail/api:", log.Lshortfile),
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

func (a *Api) SaveUser(in io.Reader) (*users.User, error) {

	raw := users.DecodeRaw(in)

	a.logger.Printf("lets check %#v", raw)
	if raw.Valid() {
		savedUser := raw.NewUser()
		a.logger.Printf("dup check save! %#v", savedUser.Email)
		exists, _ := a.userStore.GetUserByEmail(savedUser.Email)
		if exists != nil {

		}
		a.logger.Printf("lets save! %#v", savedUser)
		err := a.userStore.AddUser(savedUser)
		if err != nil {
			a.logger.Println(err.Error())
			return savedUser, ErrInternalServer.Error
		}
		return savedUser, err
	}
	a.logger.Println(ErrInvalidSignup.Error)
	return nil, ErrInvalidSignup.Error
}

func (a *Api) GetUser(id string) (*users.User, error) {
	if id == "" {
		a.logger.Println(ErrNoUserId.Error)
		return nil, ErrNoUserId.Error
	}
	foundUser, err := a.userStore.GetUser(id)
	if err != nil {
		a.logger.Println(err.Error())
		return nil, ErrInternalServer.Error
	}
	return foundUser, nil
}

func (a *Api) AuthenticateUserSession(sessionToken string) (*users.User, error) {

	a.logger.Println("token ", sessionToken)
	valid, data := users.SessionValid(sessionToken, a.Secret)
	a.logger.Printf("data %#v", data)

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
		a.logger.Println(err.Error())
		return ""
	}
	return sessionUser.SessionRefresh(sessionToken, a.Secret)
}

func (a *Api) GetUserByEmail(email string) (*users.User, error) {
	if email == "" {
		a.logger.Println(ErrNoUserId.Error)
		return nil, ErrNoUserId.Error
	}
	foundUser, err := a.userStore.GetUserByEmail(email)
	if err != nil {
		a.logger.Println(err.Error())
	}
	return foundUser, err
}

func (a *Api) SaveSmbgs(in io.Reader, out io.Writer, userid string) error {
	if userid == "" {
		a.logger.Println(ErrNoUserId.Error)
		return ErrNoUserId.Error
	}

	var dbBuffer bytes.Buffer

	smbg.StreamMulti(in, "", "", out, &dbBuffer)

	if err := a.dataStore.AddSmbgs(userid, dbBuffer.Bytes()); err != nil {
		a.logger.Println(err.Error())
		return ErrInternalServer.Error
	}
	return nil
}

func (a *Api) GetSmbgs(out io.Writer, userid string) error {
	if userid == "" {
		a.logger.Println(ErrNoUserId.Error)
		return ErrNoUserId.Error
	}

	smbgs, err := a.dataStore.GetSmbgs(userid)
	if err != nil {
		a.logger.Println(err.Error())
		return ErrInternalServer.Error
	}

	out.Write(smbgs)
	return nil
}
