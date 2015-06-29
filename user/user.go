package user

import (
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jameskeane/bcrypt"
	"github.com/satori/go.uuid"
)

const (
	secret = "WOW,MuchShibe,ToDogge"
)

// * When a user logs in to your site via a POST under TLS, determine if the password is valid.
// * Then issue a random session key, say 50 or more crypto rand characters and stuff in a secure Cookie.
// * Add that session key to the UserSession table.
// * Then when you see that user again, first hit the UserSession table to see if the SessionKey is in there with a valid LoginTime and LastSeenTime and User is not deleted. You could design it so a timer automatically clears out old rows in UserSession."

type User struct {
	Id       string    `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Hash     string    `json:"hash"`
	LastSeen time.Time `json:"lastSeen"`
	Disabled bool      `json:"disabled"`
}

//incoming raw user data used to generate a new user from
type RawUser struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Pass  string `json:"password"`
}

//incoming raw user data used to generate a new user from
type TokenData struct {
	UserId string
}

/*func NewUser(name, email, pw string) *User {
	u := &User{Id: uuid.NewV4().String(), Name: name, Email: email, Hash: pw}
	u.Encrypt()
	return u
}*/

func DecodeRaw(src io.Reader) *RawUser {

	dec := json.NewDecoder(src)
	u := &RawUser{}
	if err := dec.Decode(&u); err != nil {
		log.Println("RawUser.Decode", err.Error())
	}
	return u
}

func (u *RawUser) Valid() bool {
	return u.Email != "" && u.Name != "" && u.Pass != ""
}

func (u *RawUser) NewUser() *User {
	nu := &User{Id: uuid.NewV4().String(), Name: u.Name, Email: u.Email, Hash: u.Pass}
	nu.Encrypt()
	return nu
}

func (u *User) Encrypt() {
	u.Hash, _ = bcrypt.Hash(u.Hash)
}

func (u User) Validate(p string) bool {
	return bcrypt.Match(p, u.Hash)
}

func (u *User) Json() []byte {
	data, _ := json.Marshal(&u)
	return data
}

func (u *User) Login() string {
	//log.Println("Login")

	if u.Disabled {
		return ""
	}
	//set last seen
	u.LastSeen = time.Now().UTC()
	//create a token for this user
	token := jwt.New(jwt.GetSigningMethod("HS256"))
	token.Claims["ID"] = u.Id
	token.Claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		log.Println("Login err", err.Error())
	}
	return tokenString

}

func (u *User) SessionRefresh(sessionToken string) string {
	if u.Disabled {
		return ""
	}
	if current := unpackToken(sessionToken); current != nil && current.Valid {
		//set last seen
		u.LastSeen = time.Now().UTC()
		//create new token
		token := jwt.New(jwt.GetSigningMethod("HS256"))
		token.Claims["ID"] = current.Claims["ID"]
		token.Claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
		tokenString, err := token.SignedString([]byte(secret))
		if err != nil {
			log.Println("SessionRefresh err", err.Error())
		}
		return tokenString
	}
	return ""
}

func SessionValid(sessionToken string) (bool, *TokenData) {
	if current := unpackToken(sessionToken); current != nil {
		return current.Valid, &TokenData{UserId: current.Claims["id"].(string)}
	}
	return false, nil
}

//check and if valid return the token after setting the `LastSeen` on the associated User
func unpackToken(tokenString string) *jwt.Token {
	current, err := jwt.Parse(tokenString, func(t *jwt.Token) ([]byte, error) {
		return []byte(secret), nil
	})
	if err != nil {
		log.Println("error unpacking token:", err.Error())
		return nil
	}
	return current
}

/*

https://github.com/go-authboss/authboss-sample

type User struct {
	ID   int
	Name string

	// Auth
	Email    string
	Password string

	// OAuth2
	Oauth2Uid      string
	Oauth2Provider string
	Oauth2Token    string
	Oauth2Refresh  string
	Oauth2Expiry   time.Time

	// Confirm
	ConfirmToken string
	Confirmed    bool

	// Lock
	AttemptNumber int
	AttemptTime   time.Time
	Locked        time.Time

	// Recover
	RecoverToken       string
	RecoverTokenExpiry time.Time

	// Remember is in another table
}
*/
