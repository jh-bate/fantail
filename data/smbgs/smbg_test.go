package smbgs

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

func Test_NewSmbg(t *testing.T) {

	bg := NewSmbg()

	if bg.EventType != data.EventTypes.Smbg.String() {
		t.Fatalf("NewSmbg: expected [%s] got [%s] ", data.EventTypes.Smbg.String(), bg.EventType)
	}

	if bg.Id == "" {
		t.Fatal("NewSmbg: expects and Id is set on creation")
	}

	if errors := bg.Validate(); len(errors) == 0 {
		t.Fatal("Validate: for a brand new BloodGlucose should not be valid ")
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

	const test_file = "StreamNewSmbgs_test.json"

	const bgsStream = `
	{"value":9.9, "time": "2015-06-10T01:42:19.419Z"}
	{"value":8.9, "time": "2015-06-10T02:42:19.419Z"}
	{"value":7.9, "time": "2015-06-10T04:42:19.419Z"}
	{"value":6.9, "time": "2015-06-10T06:42:19.419Z"}
	`

	rec := httptest.NewRecorder()

	f := createFile(test_file)
	defer f.Close()

	StreamNew(strings.NewReader(bgsStream), DEVICE_ID, UPLOAD_ID, f, rec)

	testFile, err := ioutil.ReadFile(test_file)
	if err != nil {
		t.Fatalf("error [%s] reading file", err.Error())
	}

	//File
	var bgs_2 Smbgs
	json.Unmarshal(testFile, &bgs_2)

	if len(bgs_2) != 4 {
		t.Fatalf("expected [4] got [%b] ", len(bgs_2))
	}

	if bgs_2[2].Value != 7.9 {
		t.Fatalf("expected [7.9] got [%b] ", bgs_2[2].Value)
	}

	//Response
	if rec.Body.Len() != 0 {
		// compare bodies by comparing the unmarshalled JSON results
		var result interface{}
		if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
			t.Fatalf("Err decoding nonempty response body: [%v]\n [%v]\n", err, rec.Body)
		}

	} else {
		t.Fatal("no data return from StreamNewBloodGlucoses")
	}

}
func Test_Stream_Transport(t *testing.T) {
	var in bytes.Buffer // Stand-in for the network.
	var out bytes.Buffer

	inBgs := Smbgs{
		NewSmbg().setBasics(DEVICE_ID, UPLOAD_ID, nil),
		NewSmbg().setBasics(DEVICE_ID, "UPLlod_other", nil),
	}

	//i.e as a byte stream
	inBgs.encode(&in)

	StreamExisting(&in, &out)

	outBgs := decodeExisting(&out)

	if len(outBgs) == 2 {

		for i := range outBgs {
			if outBgs[i].UploadId != inBgs[i].UploadId {
				t.Fatalf("expected[%v] actual[%v]", inBgs[i].UploadId, outBgs[i].UploadId)
			}
		}
	} else {
		t.Fatalf(" only[%b] transported records found", out.Len())
	}
}
