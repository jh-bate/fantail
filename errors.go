package fantail

import (
	"errors"
	"net/http"

	"github.com/satori/go.uuid"
)

type DetailedError struct {
	Id     string `json:"id"`
	Status int    `json:"status"`
	Error  error  `json:"message"`
}

var (
	ErrInvalidSignup  = &DetailedError{Error: errors.New("fantail: key is invalid or of invalid type"), Status: http.StatusBadRequest}
	ErrInvalidSession = &DetailedError{Error: errors.New("fantail: invalid or expired session"), Status: http.StatusUnauthorized}
	ErrNoToken        = &DetailedError{Error: errors.New("fantail: no token present in request"), Status: http.StatusBadRequest}
	ErrNoUserId       = &DetailedError{Error: errors.New("fantail: no userid was set"), Status: http.StatusBadRequest}
	ErrInvalidRequest = &DetailedError{Error: errors.New("fantail: you are missing prerequisites for the request"), Status: http.StatusBadRequest}
	ErrInternalServer = &DetailedError{Error: errors.New("fantail: something went wrong"), Status: http.StatusInternalServerError}
	ErrLoadingConfig  = &DetailedError{Error: errors.New("fantail: could not load config"), Status: http.StatusInternalServerError}
	ErrLogin          = &DetailedError{Error: errors.New("fantail: could not login"), Status: http.StatusInternalServerError}
)

//id for tracking error to logged instance
func (d *DetailedError) SetId() *DetailedError {
	d.Id = uuid.NewV4().String()
	return d
}
