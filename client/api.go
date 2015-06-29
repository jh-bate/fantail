package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/jh-bate/fantail/models/smbg"

	"github.com/jh-bate/fantail/user"
)

type Api struct {
	store *Store
}

type detailedError struct {
	Id     string `json:"id"`
	Status int    `json:"status"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

// to give a better sense of what and where something went wrong
func DetailedError(id string, status int, title, detail string) error {
	detailed := detailedError{Id: id, Status: status, Title: title, Detail: detail}
	b, _ := json.Marshal(detailed)
	return errors.New(string(b))
}

var (
	ErrBadRequest     = DetailedError("bad_request", http.StatusBadRequest, "Bad request", "Request body is not well-formed. It must be JSON.")
	ErrUserSignup     = DetailedError("invalid_user_data", http.StatusBadRequest, "Bad request", "Please check that data you used to signup")
	ErrNoUserid       = DetailedError("no_userid", http.StatusBadRequest, "No userid", "The userid must be set.")
	ErrInternalServer = DetailedError("internal_server_error", http.StatusInternalServerError, "Internal Server Error", "Something went wrong.")
)

func InitApi(s *Store) *Api { return &Api{store: s} }

func (a *Api) SaveUser(src io.Reader) (*user.User, error) {

	raw := user.DecodeRaw(src)

	log.Printf("lets check %#v", raw)
	if raw.Valid() {

		usr := raw.NewUser()
		log.Printf("lets save! %#v", usr)
		err := a.store.AddUser(usr)
		return usr, err
	}
	return nil, ErrUserSignup
}

func (a *Api) GetUser(id string) (*user.User, error) {
	if id == "" {
		return nil, ErrNoUserid
	}
	return a.store.GetUser(id)
}

func (a *Api) AuthenticateUserSession(sessionToken string) (*user.User, error) {
	valid, data := user.SessionValid(sessionToken)
	if valid {
		user, err := a.GetUser(data.UserId)
		if err != nil {
			return nil, errors.New("could not find session user")
		}
		return user, nil
	}
	return nil, errors.New("invalid or expired session")
}

func (a *Api) RefreshUserSession(sessionToken string) string {
	usr, err := a.AuthenticateUserSession(sessionToken)
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	return usr.SessionRefresh(sessionToken)
}

func (a *Api) GetUserByEmail(email string) (*user.User, error) {
	if email == "" {
		return nil, ErrNoUserid
	}
	return a.store.GetUserByEmail(email)
}

func (a *Api) SaveSmbgs2(src io.Reader, out io.Writer, userid string) error {
	if userid == "" {
		return ErrNoUserid
	}

	var dbBuffer bytes.Buffer

	smbg.StreamMulti(src, "", "", out, &dbBuffer)

	//log.Println("SaveSmbgs2 Db", string(dbBuffer.Bytes()[:]))

	if err := a.store.AddSmbgs2(userid, dbBuffer.Bytes()); err != nil {
		log.Println("api/SaveSmbgs", err.Error())
		return ErrInternalServer
	}
	return nil
}

func (a *Api) GetSmbgs2(dest io.Writer, userid string) error {
	if userid == "" {
		return ErrNoUserid
	}

	smbgs, err := a.store.GetSmbgs2(userid)
	//log.Println("GetSmbgs2", string(smbgs[:]))
	if err != nil {
		log.Println("api/GetSmbgs", err.Error())
		return ErrInternalServer
	}

	//log.Println("GetSmbgs2 Got ", string(smbgs[:]))

	dest.Write(smbgs)
	return nil
}

/*func (a *Api) SaveSmbgs(src io.Reader, userid string) (smbg.Smbgs, error) {
	if userid == "" {
		return nil, ErrNoUserid
	}

	smbgs := smbg.Decode(src).Set("", "")

	if err := a.store.AddSmbgs(userid, smbgs); err != nil {
		log.Println("api/SaveSmbgs", err.Error())
		return nil, ErrInternalServer
	}
	return smbgs, nil
}

func (a *Api) GetSmbgs(dest io.Writer, userid string) error {
	if userid == "" {
		return ErrNoUserid
	}

	smbgs, err := a.store.GetSmbgs(userid)
	if err != nil {
		log.Println("api/GetSmbgs", err.Error())
		return ErrInternalServer
	}

	err = smbgs.Encode(dest)
	if err != nil {
		log.Println("api/GetSmbgs", err.Error())
		return ErrInternalServer
	}

	return nil
}
func (a *Api) UpdateSmbgs(src io.Reader, userid string) (smbg.Smbgs, error) {
	if userid == "" {
		return nil, ErrNoUserid
	}

	smbgs := smbg.DecodeExisting(src)
	if err := a.store.AddSmbgs(userid, smbgs); err != nil {
		log.Println("api/UpdateSmbgs", err.Error())
		return nil, ErrInternalServer
	}
	return smbgs, nil
}
*/
