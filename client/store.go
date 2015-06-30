package client

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/jh-bate/fantail/models"
	"github.com/jh-bate/fantail/user"
)

type Store struct {
	logger *log.Logger
}

const (
	events_db    = "fantail_data.db"
	users_bucket = "users"
)

//store created on a per user basis
func NewStore() *Store {
	return &Store{logger: log.New(os.Stdout, "fantail:", log.Lshortfile)}
}

func (s *Store) open() *bolt.DB {
	db, err := bolt.Open(events_db, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	db.Update(func(tx *bolt.Tx) error {
		//create buckets for all types we use
		tx.CreateBucketIfNotExists([]byte(models.EventTypes.Smbg.String()))
		tx.CreateBucketIfNotExists([]byte(users_bucket))
		return nil
	})
	return db
}

func (s *Store) AddUser(usr *user.User) error {
	db := s.open()
	defer db.Close()
	s.logger.Println("Adding ...", usr.Id)
	return db.Update(func(tx *bolt.Tx) error {
		eb := tx.Bucket([]byte(users_bucket))
		return eb.Put([]byte(usr.Id), usr.Json())
	})
}

func (s *Store) GetUserByEmail(email string) (*user.User, error) {
	db := s.open()
	defer db.Close()

	var usr *user.User
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

func (s *Store) GetUser(id string) (*user.User, error) {
	db := s.open()
	defer db.Close()

	var usr *user.User

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
	s.logger.Println("found user ")
	return usr, err
}

func (s *Store) AddSmbgs(userid string, data []byte) error {

	current, _ := s.GetSmbgs(userid)

	if len(current) > 0 {
		s.logger.Println("we aleady have data for [", userid, "] so updating")
		data = append(data, current...)
	}

	db := s.open()
	defer db.Close()

	err := db.Update(func(tx *bolt.Tx) error {
		eb := tx.Bucket([]byte(models.EventTypes.Smbg.String()))
		return eb.Put([]byte(userid), data)
	})

	if err != nil {
		s.logger.Println(err.Error())
	}

	return err
}

func (s *Store) GetSmbgs(userid string) ([]byte, error) {
	db := s.open()
	defer db.Close()

	var smbgs []byte

	err := db.View(func(tx *bolt.Tx) error {
		eb := tx.Bucket([]byte(models.EventTypes.Smbg.String()))
		data := eb.Get([]byte(userid))
		if len(data) > 0 {
			smbgs = make([]byte, len(data))
			s.logger.Println("found smbgs")
			copy(smbgs, data)
			return nil
		}
		return nil
	})
	if err != nil {
		s.logger.Println(err.Error())
	}
	return smbgs, err
}
