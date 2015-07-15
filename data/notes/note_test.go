package notes

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"net/http/httptest"

	"github.com/jh-bate/fantail/data"
)

const UPLOAD_ID, DEVICE_ID = "u_123", "d_123"

func Test_NewNote(t *testing.T) {

	note := NewNote()

	if note.EventType != data.EventTypes.Note.String() {
		t.Fatalf("NewNote: expected [%s] got [%s] ", data.EventTypes.Note.String(), note.EventType)
	}

	if note.Id == "" {
		t.Fatal("NewNote: expects and Id is set on creation")
	}

	if errors := note.Validate(); len(errors) == 0 {
		t.Fatal("Validate: for a brand new Note should not be valid ")
	}
}

func createFile(name string) *os.File {
	if _, err := os.Stat(name); err == nil {
		os.Remove(name)
	}
	f, _ := os.Create(name)
	return f
}

func Test_StreamNew_ToFileAndResponse(t *testing.T) {

	const test_file = "StreamNewNotes_test.json"

	const notesStream = `
	{"creatorId":"123-321", "text": "Whoop"}
	{"creatorId":"123-321", "text": "there it is"}
	`

	rec := httptest.NewRecorder()

	f := createFile(test_file)
	defer f.Close()

	StreamNew(strings.NewReader(notesStream), DEVICE_ID, UPLOAD_ID, f, rec)

	testFile, err := ioutil.ReadFile(test_file)
	if err != nil {
		t.Fatalf("error [%s] reading file", err.Error())
	}

	//File
	var notes Notes
	json.Unmarshal(testFile, &notes)

	if len(notes) != 2 {
		t.Fatalf("expected [2] got [%b] ", len(notes))
	}

	if notes[1].Text != "there it is" {
		t.Fatalf("expected [there it is] got [%s] ", notes[1].Text)
	}

	//Response
	if rec.Body.Len() != 0 {
		// compare bodies by comparing the unmarshalled JSON results
		var result interface{}
		if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
			t.Fatalf("Err decoding nonempty response body: [%v]\n [%v]\n", err, rec.Body)
		}
	} else {
		t.Fatal("no data return from StreamNew")
	}

}

func Test_Stream_Transport(t *testing.T) {

	const textOne, textTwo = "one", "two miss a few 99 100"

	var in bytes.Buffer // Stand-in for the network.
	var out bytes.Buffer

	inNotes := Notes{
		NewNote(),
		NewNote(),
	}

	inNotes[0].Text = textOne
	inNotes[1].Text = textTwo

	//i.e as a byte stream
	inNotes.encode(&in)

	StreamExisting(&in, &out)

	outNotes := decodeExisting(&out)

	if len(outNotes) == 2 {

		for i := range outNotes {
			if outNotes[i].Text != inNotes[i].Text {
				t.Fatalf("expected[%v] actual[%v]", inNotes[i].Text, outNotes[i].Text)
			}

			if outNotes[i].CreatedAt != inNotes[i].CreatedAt {
				t.Fatalf("Created data should be the same [%s] [%s]", inNotes[i].CreatedAt, outNotes[i].CreatedAt)
			}

		}
	} else {
		t.Fatalf(" only[%b] transported records found", len(outNotes))
	}
}
