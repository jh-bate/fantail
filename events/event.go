package events

import (
	"encoding/json"
	"io"
	"log"
	"sort"
	"time"

	"github.com/satori/go.uuid"
)

type Events []*Event

type Event struct {
	//
	Id string `json:"id"`
	// The type of this event
	Type string `json:"type"`
	// An ISO8601 timestamp with a timezone offset.
	Time string `json:"time"`
	// An ISO8601 timestamp for when the items was created
	CreatedAt string `json:"createdAt,omitempty"`
	// An ISO8601 timestamp for when the item was updated
	UpdatedAt string `json:"updatedAt,omitempty"`
	//A “version” for the schema. The original schema for the type will have a schemaVersion of 0, the next modification will be 1, and so on
	SchemaVersion int `json:"schemaVersion"`
	// A flag that will indicate if the datum is valid after validation has run
	Valid bool `json:"-"`
	// The actual data for this event
	Data interface{} `json:"data"`
}

func NewEvent() *Event {
	s := &Event{CreatedAt: time.Now().UTC().Format(time.RFC3339)}
	s.setId()
	return s
}

func StreamNew(rawJson io.Reader, eventJson ...io.Writer) error {
	events := decode(rawJson)
	_, err := io.MultiWriter(eventJson...).Write(events.json())
	return err
}

func StreamExisting(eventsJson io.Reader, eventsJsonOut ...io.Writer) error {
	events := decodeExisting(eventsJson)
	_, err := io.MultiWriter(eventsJsonOut...).Write(events.json())
	return err
}

func (m *Event) setId() {
	m.Id = uuid.NewV4().String()
}

func (m *Event) json() []byte {
	asJson, _ := json.Marshal(m)
	return asJson
}

func decode(src io.Reader) Events {
	events := Events{}
	dec := json.NewDecoder(src)

	count := 0
	log.Println("streaming raw ... ")
	for {

		log.Println("count ", count)

		b := NewEvent()
		if err := dec.Decode(&b); err == io.EOF {
			break
		} else if err != nil {
			log.Println("raw events decoding", err.Error())
			break
		}
		events = append(events, b)

		count++

		log.Println("count incr ", count)
	}
	return events
}

func decodeExisting(src io.Reader) Events {
	all := Events{}
	json.NewDecoder(src).Decode(&all)
	return all
}

func (e Events) encode(dest io.Writer) error {
	return json.NewEncoder(dest).Encode(e)
}

func (e Events) json() []byte {
	asJson, _ := json.Marshal(&e)
	return asJson
}

//Sorting and filters
type ByTime Events

func (a ByTime) Len() int           { return len(a) }
func (a ByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTime) Less(i, j int) bool { return a[i].Time < a[j].Time }

func (e Events) SortByTime() {
	sort.Sort(ByTime(e))
}

func (e Event) isOnWeekday() bool {
	theTime, _ := time.Parse(time.RFC3339, e.Time)
	if theTime.Weekday() == time.Saturday || theTime.Weekday() == time.Sunday {
		return false
	}
	return true
}

func (e Events) Weekdays() Events {
	var wd Events
	for i := range e {
		if e[i].isOnWeekday() {
			wd = append(wd, e[i])
		}
	}
	return wd
}

func (e Events) Weekends() Events {
	var wd Events
	for i := range e {
		if e[i].isOnWeekday() == false {
			wd = append(wd, e[i])
		}
	}
	return wd
}

func (e Events) Type(t Type) Events {

	var eot Events
	for i := range e {

		if GetType(e[i].Type) == t {
			eot = append(eot, e[i])
		}
	}
	return eot
}
