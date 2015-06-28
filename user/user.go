package user

import (
	"encoding/json"
	"io"
	"log"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jameskeane/bcrypt"
	"github.com/satori/go.uuid"
)

const secret = "some secret for signing that needs to be config"

// * When a user logs in to your site via a POST under TLS, determine if the password is valid.
// * Then issue a random session key, say 50 or more crypto rand characters and stuff in a secure Cookie.
// * Add that session key to the UserSession table.
// * Then when you see that user again, first hit the UserSession table to see if the SessionKey is in there with a valid LoginTime and LastSeenTime and User is not deleted. You could design it so a timer automatically clears out old rows in UserSession."

type User struct {
	Id       string    `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Hash     string    `json:"-"`
	LastSeen time.Time `json:"-"`
	Disabled bool      `json:"-"`
}

type RawUser struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Pass  string `json:"password"`
}

func NewUser(name, email, pw string) *User {
	u := &User{Id: uuid.NewV4().String(), Name: name, Email: email, Hash: pw}
	u.Encrypt()
	return u
}

func DecodeRaw(src io.Reader) *RawUser {

	dec := json.NewDecoder(src)
	u := &RawUser{}
	if err := dec.Decode(&u); err != nil {
		log.Println("RawUser.Decode", err.Error())
	}
	return u
}

func (u *User) Encrypt() {
	u.Hash, _ = bcrypt.Hash(u.Hash)
}

func (u User) Validate(p string) bool {
	return bcrypt.Match(p, u.Hash)
}

func (u *User) Signup(s *Store) bool {
	if savedUsr := s.GetUser(u.Email); savedUsr == nil {
		s.AddOrUpdateUser(*u)
		return true
	}
	return false
}

func (u *User) Login(s *Store) string {
	if savedUsr := s.GetUser(u.Email); savedUsr != nil {
		log.Println("create token")
		token := jwt.New(jwt.GetSigningMethod("RS256"))
		token.Claims["ID"] = u.Email
		token.Claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
		tokenString, _ := token.SignedString([]byte(secret))
		log.Println("token", tokenString)
		return tokenString
	}
	return ""
}

func SessionRefresh(givenToken string, s *Store) string {
	if current := parseToken(givenToken, s); current != nil {
		token := jwt.New(jwt.GetSigningMethod("RS256"))
		token.Claims["ID"] = current.Claims["ID"]
		token.Claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
		tokenString, _ := token.SignedString([]byte(secret))
		return tokenString
	}
	return ""
}

func SessionValid(givenToken string, s *Store) bool {
	if current := parseToken(givenToken, s); current != nil {
		return true
	}
	return false
}

//check and if valid return the token after setting the `LastSeen` on the associated User
func parseToken(token string, s *Store) *jwt.Token {
	current, err := jwt.Parse(token, func(t *jwt.Token) ([]byte, error) {
		return []byte(secret), nil
	})
	if err == nil && current.Valid == true {
		user := s.GetUser(current.Claims["ID"].(string))
		user.LastSeen = time.Now().UTC()
		s.AddOrUpdateUser(*user)
		return current
	}
	return nil
}

type Store struct {
	sync.RWMutex
	Users map[string]*User
}

func (s *Store) AddOrUpdateUser(u User) {
	s.Lock()
	s.Users[u.Email] = &u
	s.Unlock()
}

func (s *Store) GetUser(email string) *User {
	s.RLock()
	var user *User
	if s.Users[email] != nil {
		user = &User{}
		*user = *s.Users[email]
	}
	s.RUnlock()
	return user
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
