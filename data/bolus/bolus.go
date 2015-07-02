package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

type Bolus struct {
	Common
	SubType string  `json:"subType"`
	Value   float64 `json:"value"`
	Insulin string  `json:"insulin"`
}

type Boluses []*Bolus

var ErrorBolusValueNotSpecified = errors.New(fmt.Sprint(EventTypes.Bolus.String(), " Value not specified"))
var ErrorBolusInsulinNotSpecified = errors.New(fmt.Sprint(EventTypes.Bolus.String(), "  Insulin not specified"))
var ErrorBolusJsonInvalid = errors.New(fmt.Sprint(EventTypes.Bolus.String(), " JSON invalid"))

func NewBolus() *Bolus {
	return &Bolus{Common: Common{EventType: EventTypes.Bolus.String()}}
}

func (m *Bolus) Validate() (errors []error) {
	switch {
	case m.Value <= 0:
		errors = append(errors, ErrorBolusValueNotSpecified)
	case m.Insulin == "":
		errors = append(errors, ErrorBolusInsulinNotSpecified)
	}
	if commonErrors := m.Common.validate(); commonErrors != nil {
		for i := range commonErrors {
			errors = append(errors, commonErrors[i])
		}
	}

	return errors
}

func (m *Bolus) FromJSON(rawJson []byte, deviceId, uploadId string) error {
	if err := json.Unmarshal(rawJson, &m); err != nil {
		log.Print(m.EventType, " error ", err.Error())
		return ErrorBolusJsonInvalid
	}
	m.Common.set(EventTypes.Bolus, deviceId, uploadId, rawJson)
	return nil
}

func (m Boluses) JSON() []byte {
	bolusesJSON, _ := json.Marshal(m)
	return bolusesJSON
}

func NewBoluses(rawJSON []byte, deviceId, uploadId string) (Boluses, error) {

	var vals []interface{}
	if err := json.Unmarshal(rawJSON, &vals); err != nil {
		log.Print("BolusesFromJSON: ", ErrorBolusJsonInvalid.Error())
		return nil, ErrorBolusJsonInvalid
	}

	var boluses Boluses

	for i := range vals {
		if jsonBolus, err := json.Marshal(vals[i]); err == nil {
			bolus := NewBolus()
			bolus.FromJSON(jsonBolus, deviceId, uploadId)
			boluses = append(boluses, bolus)
		}
	}

	return boluses, nil
}
