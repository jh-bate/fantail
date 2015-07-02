package users

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/boltdb/bolt"
)

type Store struct {
	logger *log.Logger
	path   string
}

const (
	users_bucket = "users"
)

//store created on a per user basis
func NewStore(storePath string) *Store {
	if storePath == "" {
		log.Panic("need the path of where the data will be stored")
	}
	return &Store{logger: log.New(os.Stdout, "fantail:", log.Lshortfile), path: storePath}
}

func (s *Store) open() *bolt.DB {

	db, err := bolt.Open(s.path, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	db.Update(func(tx *bolt.Tx) error {
		//create buckets for all types we use
		tx.CreateBucketIfNotExists([]byte(users_bucket))
		return nil
	})
	return db
}

func (s *Store) AddUser(usr *User) error {
	db := s.open()
	defer db.Close()
	s.logger.Println("Adding ...", usr.Id)
	return db.Update(func(tx *bolt.Tx) error {
		eb := tx.Bucket([]byte(users_bucket))
		return eb.Put([]byte(usr.Id), usr.Json())
	})
}

func (s *Store) GetUserByEmail(email string) (*User, error) {
	db := s.open()
	defer db.Close()

	var usr *User
	s.logger.Println("Finding ...", email)
	err := db.View(func(tx *bolt.Tx) error {
		ub := tx.Bucket([]byte(users_bucket))
		c := ub.Cursor()
		// try and match
		for k, v := c.First(); k != nil; k, v = c.Next() {

			json.Unmarshal(v, &usr)
			if strings.ToLower(usr.Email) == strings.ToLower(email) {
				return nil
			}
		}
		s.logger.Println("No match found for ", email)
		//no match found
		usr = nil
		return nil
	})
	if err != nil {
		s.logger.Println(err.Error())
	}
	s.logger.Printf("Found user")
	return usr, err
}

func (s *Store) GetUser(id string) (*User, error) {
	db := s.open()
	defer db.Close()

	var usr *User

	err := db.View(func(tx *bolt.Tx) error {
		eb := tx.Bucket([]byte(users_bucket))
		data := eb.Get([]byte(id))
		if len(data) > 0 {
			return json.Unmarshal(data, &usr)
		}
		s.logger.Println("boo no user!")
		return nil
	})

	if err != nil {
		s.logger.Println(err.Error())
	}
	return usr, err
}
