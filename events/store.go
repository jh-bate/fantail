package events

import (
	"log"
	"os"
	"time"

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

//bucket per user
//backet per `upload` for that user
func (s *Store) AddEvents(userid string, data []byte) error {

	db := s.open()
	defer db.Close()

	err := db.Update(func(tx *bolt.Tx) error {
		s.logger.Println("getting data bucket")
		eb := tx.Bucket([]byte(data_bucket)) //data
		s.logger.Println("getting user bucket for ", userid)
		ub, err := eb.CreateBucketIfNotExists([]byte(userid)) //user bucket
		if err != nil {
			s.logger.Println("failed getting  ", userid)
			return err
		}
		addedDate := time.Now().UTC().String()
		s.logger.Println("add upload", addedDate, "for", userid)
		return ub.Put([]byte(addedDate), data)
	})

	if err != nil {
		s.logger.Println(err.Error())
	}

	return err
}

func (s *Store) GetEvents(userid string) ([]byte, error) {
	db := s.open()
	defer db.Close()
	events := make([]byte, 0)

	err := db.View(func(tx *bolt.Tx) error {
		eb := tx.Bucket([]byte(data_bucket)) //data
		ub := eb.Bucket([]byte(userid))      //user buckect

		ub.ForEach(func(uploadId, uploadData []byte) error {
			if len(uploadData) > 0 {
				s.logger.Println("found upload", string(uploadId), "for", userid)
				events = append(events, uploadData...)
				//copy(events, uploadData)
			}
			return nil
		})
		return nil
	})
	if err != nil {
		s.logger.Println(err.Error())
	}
	return events, err
}
