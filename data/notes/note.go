package notes

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/jh-bate/fantail/data"
)

type Note struct {
	data.Common
	//who wrote the note
	CreatorId string `json:"creatorId"`
	//the content of the note
	Text string `json:"text"`
}

type Notes []*Note

var ErrorNoteCreatorNotSpecified = errors.New(fmt.Sprint(data.EventTypes.Note.String(), " creatorId not specified"))
var ErrorNoteTextNotSpecified = errors.New(fmt.Sprint(data.EventTypes.Note.String(), " text not specified"))

func noteJsonError(method string, jsonError error) error {
	return fmt.Errorf("%s.%s: %s", data.EventTypes.Note.String(), method, jsonError.Error())
}

func NewNote() *Note {
	s := &Note{Common: data.Common{EventType: data.EventTypes.Note.String(), CreatedAt: time.Now().UTC().Format(time.RFC3339)}}
	s.SetId()
	return s
}

func (m *Note) Validate() (errors []error) {
	if m.CreatorId == "" {
		errors = append(errors, ErrorNoteCreatorNotSpecified)
	}
	if m.Text == "" {
		errors = append(errors, ErrorNoteTextNotSpecified)
	}

	if commonErrors := m.Common.Validate(); commonErrors != nil {
		for i := range commonErrors {
			errors = append(errors, commonErrors[i])
		}
	}

	//update the model based on validation results
	if len(errors) == 0 {
		m.Valid = true
	}

	return errors
}

func (m *Note) json() []byte {
	asJson, _ := json.Marshal(m)
	return asJson
}

//stream incoming data and create and then write Notes as JSON to all destinations
func StreamNew(rawJson io.Reader, deviceId, uploadId string, noteJson ...io.Writer) error {
	notes := decode(rawJson)
	mw := io.MultiWriter(noteJson...)
	_, err := mw.Write(notes.json())
	return err
}

//stream incoming existing notes and then write the values as JSON to all destinations
func StreamExisting(noteJson io.Reader, noteJsonOut ...io.Writer) error {
	all := decodeExisting(noteJson)
	mw := io.MultiWriter(noteJsonOut...)
	_, err := mw.Write(all.json())
	return err
}

func decode(src io.Reader) Notes {
	notes := Notes{}
	dec := json.NewDecoder(src)

	count := 0
	log.Println(data.EventTypes.Note.String(), "streaming raw ... ")
	for {

		log.Println("count ", count)

		n := NewNote()
		if err := dec.Decode(&n); err == io.EOF {
			break
		} else if err != nil {
			log.Println(noteJsonError("Notes.Decode", err).Error())
			break
		}
		notes = append(notes, n)

		count++

		log.Println("count incr ", count)
	}
	return notes
}

func decodeExisting(src io.Reader) Notes {
	all := Notes{}
	json.NewDecoder(src).Decode(&all)
	all.Validate()
	return all
}

func (m Notes) encode(dest io.Writer) error {
	return json.NewEncoder(dest).Encode(m)
}

func (m Notes) json() []byte {
	asJson, _ := json.Marshal(&m)
	return asJson
}

func (m Notes) Validate() Notes {
	for i := range m {
		m[i].Validate()
	}
	return m
}

func (m Notes) set(deviceId, uploadId string) Notes {
	for i := range m {
		m[i].UploadId = uploadId
		m[i].DeviceId = deviceId
		m[i].Validate()
	}
	return m
}
