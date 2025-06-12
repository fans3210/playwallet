package errs

import (
	"errors"
	"strings"
)

var (
	ErrInvalidParam        = errors.New("invalid params")
	ErrInvalidPlayer       = errors.New("invalid player")
	ErrNotFound            = errors.New("not found")
	ErrNotAllowed          = errors.New("not allowed")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrDuplicate           = errors.New("record already exist")
)

type ValidationError struct {
	Reasons []string
}

func (err ValidationError) Error() string {
	if len(err.Reasons) <= 0 {
		return ""
	}
	return strings.Join(err.Reasons, ",")
}

func ValidationErrWithReason(reason string) error {
	return ValidationError{Reasons: []string{reason}}
}
