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
	ErrStateError              = errors.New("state error")
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

	default:
		return cfg.Errors.IncorrectInput
	}
}
