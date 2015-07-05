package data

import (
	"testing"

	"github.com/jh-bate/fantail/data/smbg"
)

func TestStore(t *testing.T) {

	const (
		upload_one = "u_1"
		upload_two = "u_2"
		device_id  = "test_123_bg"
	)

	bgs := data.BloodGlucoses{
		smbg.NewSmbg().SetCommon(device_id, upload_one, nil),
		smbg.NewSmbg().SetCommon(device_id, upload_two, nil),
	}

	key := "/data/123/smbg"
	s := NewStore()

	if err := s.Put(key, bgs); err != nil {
		t.Fatal("TestStore: Failed Save ", err.Error())
	}

	var storedBgs smbg.Smbgs

	err := s.Get(key, &storedBgs)
	if err != nil {
		t.Fatal("TestStore: Failed Find ", err.Error())
	}

	if bgs[0].EventType != storedBgs[0].EventType {
		t.Fatalf("TestStore: expected [%s] got [%s] ", bgs[0].EventType, storedBgs[0].EventType)
	}

}
