package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

type Note struct {
	Common
	CreatorId string `json:"creatorId"`
	Text      string `json:"text"`
}

type Notes []*Note

var ErrorNoteCreatorNotSpecified = errors.New(fmt.Sprint(EventTypes.Note.String(), " creatorId not specified"))
var ErrorNoteJsonInvalid = errors.New(fmt.Sprint(EventTypes.Note.String(), " JSON invalid"))

func NewNote() *Note {
	return &Note{Common: Common{EventType: EventTypes.Note.String()}}
}

func (m *Note) Validate() (errors []error) {
	switch {
	case m.CreatorId == "":
		errors = append(errors, ErrorNoteCreatorNotSpecified)
	}
	if commonErrors := m.Common.validate(); commonErrors != nil {
		for i := range commonErrors {
			errors = append(errors, commonErrors[i])
		}
	}

	return errors
}

func (m *Note) FromJSON(rawJson []byte, deviceId, uploadId string) error {
	if err := json.Unmarshal(rawJson, &m); err != nil {
		log.Print(m.EventType, " error ", err.Error())
		return ErrorNoteJsonInvalid
	}
	m.Common.set(EventTypes.Note, deviceId, uploadId, rawJson)
	return nil
}

func (m Notes) JSON() []byte {
	notesJSON, _ := json.Marshal(m)
	return notesJSON
}

func NewNotes(rawJSON []byte, deviceId, uploadId string) (Notes, error) {

	var vals []interface{}
	if err := json.Unmarshal(rawJSON, &vals); err != nil {
		log.Print("NewNotes: ", ErrorNoteJsonInvalid.Error())
		return nil, ErrorNoteJsonInvalid
	}

	var notes Notes

	for i := range vals {
		if jsonNote, err := json.Marshal(vals[i]); err == nil {
			note := NewNote()
			note.FromJSON(jsonNote, deviceId, uploadId)
			notes = append(notes, note)
		}
	}

	return notes, nil
}
