package client

import (
	"testing"

	"github.com/jh-bate/d-data-cli/models"
)

func TestStore(t *testing.T) {

	const (
		upload_one = "u_1"
		upload_two = "u_2"
		device_id  = "test_123_bg"
	)

	bgs := models.BloodGlucoses{
		models.NewBloodGlucose().SetCommon(device_id, upload_one, nil),
		models.NewBloodGlucose().SetCommon(device_id, upload_two, nil),
	}

	key := "/data/123/smbg"
	s := NewStore()

	if err := s.Put(key, bgs); err != nil {
		t.Fatal("TestStore: Failed Save ", err.Error())
	}

	var storedBgs models.BloodGlucoses

	err := s.Get(key, &storedBgs)
	if err != nil {
		t.Fatal("TestStore: Failed Find ", err.Error())
	}

	if bgs[0].EventType != storedBgs[0].EventType {
		t.Fatalf("TestStore: expected [%s] got [%s] ", bgs[0].EventType, storedBgs[0].EventType)
	}

}
