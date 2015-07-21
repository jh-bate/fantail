package users

import (
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jameskeane/bcrypt"
	"github.com/satori/go.uuid"
)

// * When a user logs in to your site via a POST under TLS, determine if the password is valid.
// * Then issue a random session key, say 50 or more crypto rand characters and stuff in a secure Cookie.
// * Add that session key to the UserSession table.
// * Then when you see that user again, first hit the UserSession table to see if the SessionKey is in
//   there with a valid LoginTime and LastSeenTime and User is not deleted. You could design it so a timer
//   automatically clears out old rows in UserSession."

//user we use internally and store
type User struct {
	Id       string    `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Hash     string    `json:"hash"`
	LastSeen time.Time `json:"lastSeen"`
	Disabled bool      `json:"disabled"`
}

//user that can be sent out to the world
type PublishedUser struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
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

const (
	token_id_claim       = "ID"
	token_signing_method = "HS256"
	token_expiry_claim   = "exp"

	FANTAIL_SESSION_TOKEN = "x-fantail-token"
)

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

func (u *User) ToPublishedUser() *PublishedUser {
	return &PublishedUser{Id: u.Id, Email: u.Email, Name: u.Name}
}

func (u *User) Login(secret string) string {

	if u.Disabled {
		return ""
	}
	//set last seen
	u.LastSeen = time.Now().UTC()
	//create a token for this user
	token := jwt.New(jwt.GetSigningMethod(token_signing_method))
	token.Claims[token_id_claim] = u.Id
	token.Claims[token_expiry_claim] = time.Now().Add(time.Hour * 72).Unix()
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		log.Println("Login err", err.Error())
	}
	return tokenString

}

func (u *User) SessionRefresh(sessionToken, secret string) string {
	if u.Disabled {
		return ""
	}
	if current := unpackToken(sessionToken, secret); current != nil && current.Valid {
		//set last seen
		u.LastSeen = time.Now().UTC()
		//create new token
		token := jwt.New(jwt.GetSigningMethod(token_signing_method))
		token.Claims[token_id_claim] = current.Claims[token_id_claim]
		token.Claims[token_id_claim] = time.Now().Add(time.Hour * 72).Unix()
		tokenString, err := token.SignedString([]byte(secret))
		if err != nil {
			log.Println("SessionRefresh err", err.Error())
		}
		return tokenString
	}
	return ""
}

func SessionValid(sessionToken, secret string) (bool, *TokenData) {
	if current := unpackToken(sessionToken, secret); current != nil {
		return current.Valid, &TokenData{UserId: current.Claims[token_id_claim].(string)}
	}
	return false, nil
}

//check and if valid return the token after setting the `LastSeen` on the associated User
func unpackToken(tokenString, secret string) *jwt.Token {
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
