package data

import (
	"errors"
	"fmt"
	"log"

	"github.com/satori/go.uuid"
)

type EventType string

var EventTypes = struct {
	Smbg    EventType
	Cbg     EventType
	Food    EventType
	Note    EventType
	Basal   EventType
	Keytone EventType
	Bolus   EventType
	Unknown EventType
}{"smbg", "cbg", "food", "note", "basal", "keytone", "bolus", "unknown"}

func GetEventType(eventType string) EventType {
	switch eventType {
	case EventTypes.Smbg.String():
		return EventTypes.Smbg
	case EventTypes.Cbg.String():
		return EventTypes.Cbg
	case EventTypes.Food.String():
		return EventTypes.Food
	case EventTypes.Note.String():
		return EventTypes.Note
	case EventTypes.Basal.String():
		return EventTypes.Basal
	case EventTypes.Keytone.String():
		return EventTypes.Keytone
	case EventTypes.Bolus.String():
		return EventTypes.Bolus
	}
	return EventTypes.Unknown
}

func (t EventType) String() string {
	switch t {
	case EventTypes.Smbg:
		return "smbg"
	case EventTypes.Cbg:
		return "cbg"
	case EventTypes.Food:
		return "food"
	case EventTypes.Note:
		return "note"
	case EventTypes.Basal:
		return "basal"
	case EventTypes.Keytone:
		return "keytone"
	case EventTypes.Bolus:
		return "basal"
	}
	return "unknown"
}

type Common struct {
	//
	Id string `json:"id"`
	// The type of this event
	EventType string `json:"type"`
	// An indication of the device that generated the datum. This should be globally unique to this device and repeatable with each upload.
	// A device make and model with serial number, shortened, is a good value to include here.
	DeviceId string `json:"deviceId,omitempty"`
	// The upload identifier; this field should be the uploadId of the corresponding upload data record.
	UploadId string `json:"uploadId,omitempty"`
	// An ISO8601 timestamp with a timezone offset.
	Time string `json:"time"`
	// An ISO8601 timestamp for when the items was created
	CreatedAt string `json:"createdAt,omitempty"`
	// An ISO8601 timestamp for when the item was updated
	UpdatedAt string `json:"updatedAt,omitempty"`
	//A “version” for the datum. The original datum will have a datumVersion of 0, the next modification will be 1, and so on
	DatumVersion int `json:"datumVersion"`
	//A “version” for the schema. The original schema for the type will have a schemaVersion of 0, the next modification will be 1, and so on
	SchemaVersion int `json:"schemaVersion"`
	//A flag that will indicate if the datum is valid after validation has run
	Valid bool `json:"-"`
	// Any object can have a “payload”, which is itself an object with any number of unspecified fields.
	// This is used to store device-specific data that doesn’t rise to the level of standard, but that is useful to store.
	// For example, the Dexcom G4 continuous glucose monitor displays “trend arrows”, which are arrows that indicate the general direction of the change in glucose readings – up, down, flat. This information is stored under payload.trend.
	Payload interface{} `json:"payload,omitempty"`
}

// Error's that related to the common fields after validation
var ErrorTypeNotSpecified = errors.New("type not specified")
var ErrorDeviceIdNotSpecified = errors.New("deviceId not specified")
var ErrorTimeNotSpecified = errors.New("time not specified")
var ErrorIdNotSet = errors.New("id not set")

func (m *Common) Validate() (errs []error) {
	if m.Id == "" {
		errs = append(errs, ErrorIdNotSet)
	}
	if m.EventType == "" {
		errs = append(errs, ErrorTypeNotSpecified)
	}
	if m.DeviceId == "" {
		errs = append(errs, ErrorDeviceIdNotSpecified)
	}
	if m.Time == "" {
		errs = append(errs, ErrorTimeNotSpecified)
	}
	return errs
}

func (m *Common) Set(dataType EventType, deviceId, uploadId string, payload interface{}) {
	m.EventType = dataType.String()
	if m.DeviceId == "" {
		m.DeviceId = deviceId
	}
	if m.UploadId == "" {
		m.UploadId = uploadId
	}
	if m.Id == "" {
		m.SetId()
	}
	if m.DatumVersion >= 0 {
		m.DatumVersion = 0
	}
	if m.SchemaVersion >= 0 {
		m.SchemaVersion = 0
	}
	m.Payload = payload
	m.Valid = false //false by default
	if errs := m.Validate(); len(errs) == 0 {
		m.Valid = true
	}
}

func (m *Common) SetId() {
	m.Id = uuid.NewV4().String()
}

func ErrorReport(errors []error) {
	rpt := ""

	for i := range errors {
		rpt = fmt.Sprintln(errors[i])
	}

	log.Print(rpt)
}
