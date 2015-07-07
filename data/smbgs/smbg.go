package smbgs

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/jh-bate/fantail/data"
)

// SMBG represents blood glucose from a finger prick or other “self-monitoring” method. These events are point-in-time and look like
type Smbg struct {
	data.Common
	// the bloodglucose value from a self monitoring device
	Value float64 `json:"value"`
}

type Smbgs []*Smbg

var ErrorSmbgValueNotSpecified = fmt.Errorf(data.EventTypes.Smbg.String(), " Value not specified")

func smbgJsonError(method string, jsonError error) error {
	return fmt.Errorf("%s.%s: %s", data.EventTypes.Smbg.String(), method, jsonError.Error())
}

func NewSmbg() *Smbg {
	s := &Smbg{Common: data.Common{EventType: data.EventTypes.Smbg.String(), CreatedAt: time.Now().UTC().Format(time.RFC3339)}}
	s.SetId()
	return s
}

func (m *Smbg) setBasics(deviceId, uploadId string, payload interface{}) *Smbg {
	m.DeviceId = deviceId
	m.UploadId = uploadId
	m.Payload = payload
	return m
}

func (m *Smbg) Validate() (errors []error) {
	if m.Value <= 0 {
		errors = append(errors, ErrorSmbgValueNotSpecified)
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

func (m *Smbg) json() []byte {
	asJson, _ := json.Marshal(m)
	return asJson
}

//stream incoming data and create and then write Smbg values as JSON to all destinations
func StreamNew(rawJson io.Reader, deviceId, uploadId string, smbgJson ...io.Writer) error {
	smbgs := decode(rawJson).set(deviceId, uploadId)
	mw := io.MultiWriter(smbgJson...)
	_, err := mw.Write(smbgs.json())
	return err
}

//stream incoming existing smbgs and then write the values as JSON to all destinations
func StreamExisting(smbgJson io.Reader, smbgJsonOut ...io.Writer) error {
	smbgs := decodeExisting(smbgJson)
	mw := io.MultiWriter(smbgJsonOut...)
	_, err := mw.Write(smbgs.json())
	return err
}

func decode(src io.Reader) Smbgs {
	bgs := Smbgs{}
	dec := json.NewDecoder(src)

	count := 0
	log.Println(data.EventTypes.Smbg.String(), "streaming raw ... ")
	for {

		log.Println("count ", count)

		b := NewSmbg()
		if err := dec.Decode(&b); err == io.EOF {
			break
		} else if err != nil {
			log.Println(smbgJsonError("Smbgs.Decode", err).Error())
			break
		}
		bgs = append(bgs, b)

		count++

		log.Println("count incr ", count)
	}
	return bgs
}

func decodeExisting(src io.Reader) Smbgs {
	all := Smbgs{}
	json.NewDecoder(src).Decode(&all)
	all.Validate()
	return all
}

func (bgs Smbgs) encode(dest io.Writer) error {
	return json.NewEncoder(dest).Encode(bgs)
}

func (bgs Smbgs) json() []byte {
	asJson, _ := json.Marshal(&bgs)
	return asJson
}

func (bgs Smbgs) Validate() Smbgs {
	for i := range bgs {
		bgs[i].Validate()
	}
	return bgs
}

func (bgs Smbgs) set(deviceId, uploadId string) Smbgs {
	for i := range bgs {
		bgs[i].UploadId = uploadId
		bgs[i].DeviceId = deviceId
		bgs[i].Validate()
	}
	return bgs
}
