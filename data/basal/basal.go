package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

type Basal struct {
	Common
	DeliveryType string  `json:"deliveryType"`
	Value        float64 `json:"value"`
	Duration     int     `json:"duration"`
	Insulin      string  `json:"insulin"`
}

type Basals []*Basal

var ErrorBasalDeliveryTypeNotSpecified = errors.New(fmt.Sprint(EventTypes.Basal.String(), " deliveryType not specified"))
var ErrorBasalValueNotSpecified = errors.New(fmt.Sprint(EventTypes.Basal.String(), " value not specified"))
var ErrorBasalInsulinNotSpecified = errors.New(fmt.Sprint(EventTypes.Basal.String(), " insulin not specified"))

func basalJsonError(method string, jsonError error) error {
	return fmt.Errorf("%s.%s: %s", EventTypes.Basal.String(), method, jsonError.Error())
}

func NewBasal() *Basal {
	return &Basal{Common: Common{EventType: EventTypes.Basal.String()}}
}

func (m *Basal) Validate() (errors []error) {
	switch {
	case m.Value <= 0:
		errors = append(errors, ErrorBasalValueNotSpecified)
	case m.DeliveryType == "":
		errors = append(errors, ErrorBasalDeliveryTypeNotSpecified)
	case m.Insulin == "":
		errors = append(errors, ErrorBasalInsulinNotSpecified)
	}

	if commonErrors := m.Common.validate(); commonErrors != nil {
		for i := range commonErrors {
			errors = append(errors, commonErrors[i])
		}
	}

	return errors
}

func NewBasalFromJSON(rawJson []byte, deviceId, uploadId string) (*Basal, error) {
	m := &Basal{}
	m.Common.set(EventTypes.Basal, deviceId, uploadId, rawJson)

	if err := json.Unmarshal(rawJson, &m); err != nil {
		log.Print(m.EventType, " error ", err.Error())
		return m, basalJsonError("NewBasalFromJSON", err)
	}

	return m, nil
}

func (m Basals) JSON() []byte {
	basalsJSON, _ := json.Marshal(m)
	return basalsJSON
}

func NewBasals(rawJSON []byte, deviceId, uploadId string) (Basals, []error) {

	var errors []error

	var vals []interface{}
	if err := json.Unmarshal(rawJSON, &vals); err != nil {
		jsonErr := basalJsonError("NewBasals", err)
		log.Print("BasalsFromJSON: ", jsonErr.Error())
		errors = append(errors, jsonErr)
		return nil, errors
	}

	var basals Basals

	for i := range vals {
		if jsonBasal, err := json.Marshal(vals[i]); err == nil {
			if basal, err := NewBasalFromJSON(jsonBasal, deviceId, uploadId); err != nil {
				errors = append(errors, err) //append so we can report on errors
				log.Printf("NewBloodGlucoses: error ", err.Error())
			} else {
				basals = append(basals, basal)
			}
		}
	}

	return basals, errors
}
