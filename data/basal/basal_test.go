package models

import (
	"testing"
)

func Test_NewBasal(t *testing.T) {

	b := NewBasal()

	if b.EventType != EventTypes.Basal.String() {
		t.Fatalf("NewBasal: expected [%s] got [%s] ", EventTypes.Basal.String(), b.EventType)
	}

	if errors := b.Validate(); len(errors) == 0 {
		t.Fatal("Validate: for a brand new Basal should not be valid ")
	}

}

func Test_NewBasalFromJson(t *testing.T) {
	const (
		amount = 12
	)

	b, err := NewBasalFromJSON([]byte(`{"value":12}`), "deviceId", "uploadId")
	if err != nil {
		t.Fatalf("expected[nil] actual[%v]", err)
	}

	if b.Value != amount {
		t.Fatalf("Value: expected[%v] actual[%v]", amount, b.Value)
	}

	if b.EventType != EventTypes.Basal.String() {
		t.Fatalf("EventType: expected[%v] actual[%v]", EventTypes.Basal.String(), b.EventType)
	}

}

/*func Test_NewBasalFromJson_All(t *testing.T) {

	bg := NewBloodGlucose()

	bg.FromJSON([]byte(`{"value":9.9, "deviceId":"456", "uploadId": "123","source":"tests","time":"2015-05-28T10:40:32.572Z" }`))

	if bg.Value != 9.9 {
		t.Fatalf("Value: expected[%v] actual[%v]", 9.9, bg.Value)
	}
	if bg.Source != "tests" {
		t.Fatalf("Source: expected[%v] actual[%v]", "tests", bg.Source)
	}

	if bg.DeviceId != "456" {
		t.Fatalf("DeviceId: expected[%v] actual[%v]", "456", bg.DeviceId)
	}
	if bg.UploadId != "123" {
		t.Fatalf("UploadId: expected[%v] actual[%v]", "123", bg.UploadId)
	}
	if bg.Time != "2015-05-28T10:40:32.572Z" {
		t.Fatalf("Time: expected[%v] actual[%v]", "2015-05-28T10:40:32.572Z", bg.Time)
	}
	if bg.EventType != EventTypes.Smbg.String() {
		t.Fatalf("EventType: expected[%v] actual[%v]", EventTypes.Smbg.String(), bg.EventType)
	}
}*/

func Test_NewBasalFromJson_Invalid(t *testing.T) {

	_, err := NewBasalFromJSON([]byte(`{"wrong":12, deviceId:"456"}`), "deviceId", "uploadId")

	if err == nil {
		t.Fatal(" expected error to thrown for invalid JSON")
	}

}
