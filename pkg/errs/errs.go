package errs

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidParam        = errors.New("invalid params")
	ErrInvalidPlayer       = errors.New("invalid player")
	ErrNotFound            = errors.New("not found")
	ErrNotAllowed          = errors.New("not allowed")
	ErrInsufficientBalance = errors.New("insufficient balance")
)

type DBError struct {
	Reason string
}

func (err DBError) Error() string {
	return fmt.Sprintf("dberr: %s", err.Reason)
}

// WARN: just used to join err msgs, original err won't be wrapped
func DBErrorWithErr(err error, reasons ...string) error {
	var reason string
	if len(reasons) > 0 {
		reason = strings.Join(reasons, ",")
	}
	if err != nil {
		return DBError{fmt.Sprintf("%s,%v", reason, err)}
	}
	return DBError{reason}
}

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
