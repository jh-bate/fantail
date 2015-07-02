package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

//
type Food struct {
	Common
	Carbs float64 `json:"carbs"`
}

var ErrorFoodCarbsNotSpecified = errors.New(fmt.Sprint(EventTypes.Food.String(), " carbs not specified"))
var ErrorFoodJsonInvalid = errors.New(fmt.Sprint(EventTypes.Food.String(), " JSON invalid"))

func NewFood() *Food {
	return &Food{Common: Common{EventType: EventTypes.Food.String()}}
}

func (m *Food) FromJSON(rawJson []byte) error {
	if err := json.Unmarshal(rawJson, &m); err != nil {
		log.Print(m.EventType, " error ", err.Error())
		return ErrorFoodJsonInvalid
	}
	return nil
}

func (m *Food) Validate() (errors []error) {
	switch {
	case m.Carbs <= 0:
		errors = append(errors, ErrorFoodCarbsNotSpecified)
	}
	if commonErrors := m.Common.validate(); commonErrors != nil {
		for i := range commonErrors {
			errors = append(errors, commonErrors[i])
		}
	}

	return errors
}
