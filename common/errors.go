package common

import (
	"errors"

	"github.com/warmte/tripl/lang"
)

var (
	ErrTooLongField            = errors.New("too long field")
	ErrIncorrectDate           = errors.New("incorrect date")
	ErrIncorrectTime           = errors.New("incorrect time")
	ErrIncorrectDateTime       = errors.New("incorrect datetime")
	ErrIncorrectNumberOfFields = errors.New("incorrect number of fields")
	ErrIncorrectInput          = errors.New("incorrect input")
	ErrEmptyList               = errors.New("list is empty")
	ErrDatetimeRestraint       = errors.New("datetime restraint")
	ErrInternalError           = errors.New("internal error")
)

func GetErrorMessage(err error, cfg lang.Cfg) string {
	switch err {
	case ErrTooLongField:
		return cfg.Errors.TooLongField

	case ErrIncorrectDate:
		return cfg.Errors.IncorrectDate

	case ErrIncorrectTime:
		return cfg.Errors.IncorrectTime

	case ErrIncorrectDateTime:
		return cfg.Errors.IncorrectDatetime

	case ErrIncorrectNumberOfFields:
		return cfg.Errors.IncorrectNumberOfFields

	case ErrDatetimeRestraint:
		return cfg.Errors.DatetimeRestraint

	default:
		return cfg.Errors.InternalError
	}
}
