package events

import "errors"

var (
	ErrorIdValidation   = errors.New("id not set")
	ErrorTypeValidation = errors.New("type was not set")
	ErrorTimeValidation = errors.New("time is invalid")

	ErrorSmbgValueValidation = errors.New("smbg value not set")

	ErrorNoteTextValidation   = errors.New("note text is empty")
	ErrorNoteAuthorValidation = errors.New("note authorId is not set")
)

type Type string

var Types = struct {
	Smbg    Type
	Cbg     Type
	Food    Type
	Note    Type
	Basal   Type
	Bolus   Type
	Unknown Type
}{"smbg", "cbg", "food", "note", "basal", "bolus", "unknown"}

func GetType(Type string) Type {
	switch Type {
	case Types.Smbg.String():
		return Types.Smbg
	case Types.Cbg.String():
		return Types.Cbg
	case Types.Food.String():
		return Types.Food
	case Types.Note.String():
		return Types.Note
	case Types.Basal.String():
		return Types.Basal
	case Types.Bolus.String():
		return Types.Bolus
	}
	return Types.Unknown
}

func (t Type) String() string {
	switch t {
	case Types.Smbg:
		return "smbg"
	case Types.Cbg:
		return "cbg"
	case Types.Food:
		return "food"
	case Types.Note:
		return "note"
	case Types.Basal:
		return "basal"
	case Types.Bolus:
		return "bolus"
	}
	return "unknown"
}

type Smbg struct {
	Value float64 `json:"value"`
	Units string  `json:"units"`
}

type Note struct {
	Text     string `json:"value"`
	AuthorId string `json:"authorId"`
}

func (e Smbg) Validate() (errs []error) {
	if e.Value <= 0 {
		errs = append(errs, ErrorSmbgValueValidation)
	}
	return errs
}

func (e Note) Validate() (errs []error) {
	if e.Text == "" {
		errs = append(errs, ErrorNoteTextValidation)
	}
	if e.AuthorId == "" {
		errs = append(errs, ErrorNoteTextValidation)
	}
	return errs
}

func (e *Event) Validate() (errs []error) {
	if e.Id == "" {
		errs = append(errs, ErrorIdValidation)
	}
	if e.Type == "" {
		errs = append(errs, ErrorTypeValidation)
	}
	if e.Time == "" {
		errs = append(errs, ErrorTimeValidation)
	}
	return errs
}
