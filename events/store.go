package events

import (
	"log"
	"os"

	"github.com/boltdb/bolt"
)

type Store struct {
	logger *log.Logger
	path   string
}

const data_bucket = "data"

//EventStore created on a per user basis
func NewStore(StorePath string) *Store {
	if StorePath == "" {
		log.Panic("need the path of where the data will be EventStored")
	}
	return &Store{logger: log.New(os.Stdout, "fantail:", log.Lshortfile), path: StorePath}
}

func (s *Store) open() *bolt.DB {

	db, err := bolt.Open(s.path, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	db.Update(func(tx *bolt.Tx) error {
		//create buckets for all types we use
		tx.CreateBucketIfNotExists([]byte(data_bucket))
		return nil
	})
	return db
}

//nested bucket / userid and upload?
func (s *Store) AddEvents(userid string, data []byte) error {

	current, _ := s.GetEvents(userid)

	if len(current) > 0 {
		s.logger.Println("we aleady have data for [", userid, "] so updating")
		data = append(data, current...)
	}

	db := s.open()
	defer db.Close()

	err := db.Update(func(tx *bolt.Tx) error {
		eb := tx.Bucket([]byte(data_bucket))
		return eb.Put([]byte(userid), data)
	})

	if err != nil {
		s.logger.Println(err.Error())
	}

	return err
}

func (s *Store) GetEvents(userid string) ([]byte, error) {
	db := s.open()
	defer db.Close()

	var events []byte

	err := db.View(func(tx *bolt.Tx) error {
		eb := tx.Bucket([]byte(data_bucket))
		data := eb.Get([]byte(userid))
		if len(data) > 0 {
			events = make([]byte, len(data))
			s.logger.Println("found events")
			copy(events, data)
			return nil
		}
		return nil
	})
	if err != nil {
		s.logger.Println(err.Error())
	}
	return events, err
}
