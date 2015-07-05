package data

import (
	"log"
	"os"

	"github.com/boltdb/bolt"
)

type Store struct {
	logger *log.Logger
	path   string
}

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
		tx.CreateBucketIfNotExists([]byte(EventTypes.Smbg.String()))
		return nil
	})
	return db
}

func (s *Store) AddSmbgs(userid string, smbgsData []byte) error {

	current, _ := s.GetSmbgs(userid)

	if len(current) > 0 {
		s.logger.Println("we aleady have data for [", userid, "] so updating")
		smbgsData = append(smbgsData, current...)
	}

	db := s.open()
	defer db.Close()

	err := db.Update(func(tx *bolt.Tx) error {
		eb := tx.Bucket([]byte(EventTypes.Smbg.String()))
		return eb.Put([]byte(userid), smbgsData)
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
		eb := tx.Bucket([]byte(EventTypes.Smbg.String()))
		smbgsData := eb.Get([]byte(userid))
		if len(smbgsData) > 0 {
			smbgs = make([]byte, len(smbgsData))
			s.logger.Println("found smbgs")
			copy(smbgs, smbgsData)
			return nil
		}
		return nil
	})
	if err != nil {
		s.logger.Println(err.Error())
	}
	return smbgs, err
}
