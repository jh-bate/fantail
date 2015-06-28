package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
	"github.com/jh-bate/d-data-cli/models"

	"github.com/jh-bate/d-data-cli/models/smbg"
)

type Store struct{}

const (
	events_db    = "%s_eventdata.db"
	users_backet = "users"
)

//store created on a per user basis
func NewStore() *Store { return &Store{} }

func (s *Store) open(userid string) *bolt.DB {
	db, err := bolt.Open(fmt.Sprintf(events_db, userid), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte(models.EventTypes.Smbg.String()))
		return nil
	})
	return db
}
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

func (s *Store) AddSmbgs2(userid string, data []byte) error {

	current, _ := s.GetSmbgs2(userid)

	if len(current) > 0 {
		log.Println("we aleady have data for [", userid, "] so updating")
		data = append(data, current...)
	}

	db := s.open(userid)
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		eb := tx.Bucket([]byte(models.EventTypes.Smbg.String()))
		return eb.Put([]byte(userid), data)
	})
}

func (s *Store) GetSmbgs2(userid string) ([]byte, error) {
	db := s.open(userid)
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
