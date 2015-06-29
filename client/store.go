package client

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/jh-bate/fantail/models"
	"github.com/jh-bate/fantail/user"
)

type Store struct{}

const (
	events_db    = "fantail_data.db"
	users_bucket = "users"
)

//store created on a per user basis
func NewStore() *Store { return &Store{} }

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
	log.Println("Adding ...", usr.Id)
	//existingUsr, _ := GetUserByEmail(usr.Email)
	//if existingUsr == nil {
	log.Println("No existing user so adding ...", usr.Id)
	return db.Update(func(tx *bolt.Tx) error {
		eb := tx.Bucket([]byte(users_bucket))

		return eb.Put([]byte(usr.Id), usr.Json())
	})
	//}

	//return errors.New("user already exists")
}

func (s *Store) GetUserByEmail(email string) (*user.User, error) {
	db := s.open()
	defer db.Close()

	var usr *user.User
	log.Println("Looking for ...", email)
	err := db.View(func(tx *bolt.Tx) error {
		ub := tx.Bucket([]byte(users_bucket))
		c := ub.Cursor()
		// try and match
		for k, v := c.First(); k != nil; k, v = c.Next() {

			json.Unmarshal(v, &usr)
			log.Println("Checking ...", usr.Email)
			if strings.ToLower(usr.Email) == strings.ToLower(email) {
				return nil
			}
		}
		//no match found
		usr = nil
		return nil
	})
	log.Printf("found user %#v ", usr)
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
		log.Println("boo no user!")
		return nil
	})
	log.Printf("found user %#v ", usr)
	return usr, err
}

func (s *Store) AddSmbgs2(userid string, data []byte) error {

	current, _ := s.GetSmbgs2(userid)

	if len(current) > 0 {
		log.Println("we aleady have data for [", userid, "] so updating")
		data = append(data, current...)
	}

	db := s.open()
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		eb := tx.Bucket([]byte(models.EventTypes.Smbg.String()))
		return eb.Put([]byte(userid), data)
	})
}

func (s *Store) GetSmbgs2(userid string) ([]byte, error) {
	db := s.open()
	defer db.Close()

	var smbgs []byte

	err := db.View(func(tx *bolt.Tx) error {
		eb := tx.Bucket([]byte(models.EventTypes.Smbg.String()))
		data := eb.Get([]byte(userid))
		if len(data) > 0 {
			smbgs = make([]byte, len(data))
			//log.Println("yay we have data! ", string(data[:]))
			copy(smbgs, data)
			return nil
		}
		log.Println("boo no data!")
		return nil
	})
	//log.Println("return form db ", string(smbgs[:]))
	return smbgs, err
}

/*

func (s *Store) AddSmbgs(userid string, data smbg.Smbgs) error {

	current, _ := s.GetSmbgs(userid)

	if len(current) > 0 {
		log.Println("we aleady have data for [", userid, "] so updating")
		data = append(data, current...)
	}

	db := s.open(userid)
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		eb := tx.Bucket([]byte(models.EventTypes.Smbg.String()))
		return eb.Put([]byte(userid), data.Json())
	})
}

func (s *Store) GetSmbgs(userid string) (smbg.Smbgs, error) {
	db := s.open(userid)
	defer db.Close()

	var smbgs smbg.Smbgs

	err := db.View(func(tx *bolt.Tx) error {
		eb := tx.Bucket([]byte(models.EventTypes.Smbg.String()))
		dataBuffer := bytes.NewBuffer(eb.Get([]byte(userid)))
		if dataBuffer.Len() > 0 {
			log.Println("yay we have data!")
			smbgs = smbg.DecodeExisting(dataBuffer)
			return nil
		}
		log.Println("boo no data!")
		return nil
	})
	return smbgs, err
}

func (s *Store) Put(path string, data interface{}) error {
	db := s.open("test_123")
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		eb := tx.Bucket([]byte(models.EventTypes.Unknown.String()))
		jsonData, _ := json.Marshal(data)
		return eb.Put([]byte(path), jsonData)
	})
}

func (s *Store) Get(path string, data interface{}) error {
	db := s.open("test_123")
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		eb := tx.Bucket([]byte(models.EventTypes.Unknown.String()))
		jsonData := eb.Get([]byte(path))
		if len(jsonData) > 0 {
			return json.Unmarshal(jsonData, &data)
		}
		log.Println("get found no data ", path)
		return nil
	})
}
*/
