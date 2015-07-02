package smbg

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"net/http/httptest"
)

const UPLOAD_ID, DEVICE_ID = "u_123", "d_123"

func Test_NewSmbg(t *testing.T) {

	bg := NewSmbg()

	if bg.EventType != EventTypes.Smbg.String() {
		t.Fatalf("NewBloodGlucose: expected [%s] got [%s] ", EventTypes.Smbg.String(), bg.EventType)
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

func Test_StreamMulti_ToFileAndResponse(t *testing.T) {

	const test_file = "_StreamNewBloodGlucoses"

	const bgsStream = `
	{"value":9.9, "time": "2015-06-10T01:42:19.419Z"}
	{"value":8.9, "time": "2015-06-10T02:42:19.419Z"}
	{"value":7.9, "time": "2015-06-10T04:42:19.419Z"}
	{"value":6.9, "time": "2015-06-10T06:42:19.419Z"}
	`

	rec := httptest.NewRecorder()

	f := createFile(test_file)
	defer f.Close()

	StreamMulti(strings.NewReader(bgsStream), DEVICE_ID, UPLOAD_ID, f, rec)

	testFile, err := ioutil.ReadFile(test_file)
	if err != nil {
		t.Fatalf("error [%s] reading file", err.Error())
	}

	//File
	var bgs_2 BloodGlucoses
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
func Test_Smbgs_Transport(t *testing.T) {
	var network bytes.Buffer // Stand-in for the network.

	bgs := Smbgs{
		NewSmbg().SetCommon(DEVICE_ID, UPLOAD_ID, nil),
		NewSmbg().SetCommon(DEVICE_ID, "UPLlod_other", nil),
	}

	bgs.Encode(&network)

	otherSideBgs := DecodeExisting(&network)

	if len(otherSideBgs) == 2 {

		for i := range otherSideBgs {
			if otherSideBgs[i].UploadId != bgs[i].UploadId {
				t.Fatalf("expected[%v] actual[%v]", bgs[i].UploadId, otherSideBgs[i].UploadId)
			}
		}
	} else {
		t.Fatalf(" only[%b] transported records found", len(otherSideBgs))
	}
}
