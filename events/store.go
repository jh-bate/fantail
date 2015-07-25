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

func (s *Store) open(userid string) *bolt.DB {

	db, err := bolt.Open(s.path, 0600, nil)
	if err != nil {
		s.logger.Panic(err.Error())
	}

	err = db.Update(func(tx *bolt.Tx) error {
		//create buckets for all types we use
		eventsB, err := tx.CreateBucketIfNotExists([]byte(data_bucket))
		if err != nil {
			s.logger.Println("failed creating events bucket error:", err.Error())
			return err
		}
		//create the nested user bucket
		_, err = eventsB.Tx().CreateBucketIfNotExists([]byte(userid))
		if err != nil {
			s.logger.Println("failed creating user bucket ", userid, "error:", err.Error())
			return err
		}
		return nil
	})
	if err != nil {
		s.logger.Panic(err.Error())
	}
	return db
}

//bucket per user
//user bucket stores per `upload` for that user
func (s *Store) AddEvents(userid string, data []byte) error {

	db := s.open(userid)
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		eventsB := tx.Bucket([]byte(data_bucket)) //data
		userB := eventsB.Tx().Bucket([]byte(userid))
		addedDate := time.Now().UTC().String()
		s.logger.Println("adding upload", addedDate, "for", userid)
		return userB.Put([]byte(addedDate), data) //add events per user upload
	})
}

func (s *Store) GetEvents(userid string) ([]byte, error) {
	db := s.open(userid)
	defer db.Close()
	events := make([]byte, 0)

	err := db.View(func(tx *bolt.Tx) error {
		eventsB := tx.Bucket([]byte(data_bucket))    //data
		userB := eventsB.Tx().Bucket([]byte(userid)) //nested per user
		s.logger.Println("getting uploads for", userid)
		userB.ForEach(func(uploadId, uploadData []byte) error {
			if len(uploadData) > 0 {
				s.logger.Println("found upload", string(uploadId), "for", userid)
				uploadD := make([]byte, len(uploadData))
				copy(uploadD, uploadData)
				events = append(events, uploadD...)
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
