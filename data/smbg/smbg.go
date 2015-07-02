package smbg

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/jh-bate/d-data-cli/models"
)

// SMBG represents blood glucose from a finger prick or other “self-monitoring” method. These events are point-in-time and look like
type Smbg struct {
	models.Common
	// the bloodglucose value from a self monitoring device
	Value float64 `json:"value"`
}

type Smbgs []*Smbg

var ErrorSmbgValueNotSpecified = fmt.Errorf(models.EventTypes.Smbg.String(), " Value not specified")

func smbgJsonError(method string, jsonError error) error {
	return fmt.Errorf("%s.%s: %s", models.EventTypes.Smbg.String(), method, jsonError.Error())
}

func NewSmbg() *Smbg {
	s := &Smbg{Common: models.Common{EventType: models.EventTypes.Smbg.String(), CreatedAt: time.Now().UTC().Format(time.RFC3339)}}
	s.SetId()
	return s
}

func (m *Smbg) SetCommon(deviceId, uploadId string, payload interface{}) *Smbg {
	m.Set(models.EventTypes.Smbg, deviceId, uploadId, payload)
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

func (m *Smbg) Json() []byte {
	asJson, _ := json.Marshal(m)
	return asJson
}

//stream incoming data and create and then write Smbg values as JSON to all destinations
func StreamMulti(src io.Reader, deviceId, uploadId string, destinations ...io.Writer) error {
	smbgs := Decode(src).Set(deviceId, uploadId)
	mw := io.MultiWriter(destinations...)
	_, err := mw.Write(smbgs.Json())
	return err
}

//write incoming data to a []byte
func Write(src io.Reader, deviceId, uploadId string) []byte {
	return Decode(src).Set(deviceId, uploadId).Json()
}

func Decode(src io.Reader) Smbgs {
	bgs := Smbgs{}
	dec := json.NewDecoder(src)

	count := 0
	log.Println(models.EventTypes.Smbg.String(), "streaming ... ")
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

func DecodeExisting(src io.Reader) Smbgs {
	all := Smbgs{}
	json.NewDecoder(src).Decode(&all)
	all.Validate()
	return all
}

func (bgs Smbgs) Encode(dest io.Writer) error {
	return json.NewEncoder(dest).Encode(bgs)
}

func (bgs Smbgs) Json() []byte {
	asJson, _ := json.Marshal(&bgs)
	return asJson
}

func (bgs Smbgs) Validate() Smbgs {
	for i := range bgs {
		bgs[i].Validate()
	}
	return bgs
}

func (bgs Smbgs) Set(deviceId, uploadId string) Smbgs {
	for i := range bgs {
		bgs[i].UploadId = uploadId
		bgs[i].DeviceId = deviceId
		bgs[i].Validate()
	}
	return bgs
}
