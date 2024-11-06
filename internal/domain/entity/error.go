package entity

import (
	"runtime/debug"
)

type ErrType string

const (
	ErrTypeUnknown      ErrType = "unknown"
	ErrTypeNotFound     ErrType = "not_found"
	ErrTypeUnauthorized ErrType = "unauthorized"
	ErrTypeValidation   ErrType = "validation_error"
)

type Err struct {
	Message    string
	StackTrace string
	Type       ErrType
}

func (e Err) Error() string {
	return e.Message
}

// NewErr creates a new Err instance from either an error or a string,
// and sets the Type flag to unknown. This is useful when you want to
// create an error that is not expected to happen, and you want to
// log it with stack tracing.
func NewErr(err any) *Err {
	return newErr(err, ErrTypeUnknown)
}

func newErr(err any, errType ErrType) *Err {
	switch v := err.(type) {
	case *Err:
		return v
	case error:
		return &Err{
			Message:    v.Error(),
			StackTrace: string(debug.Stack()),
			Type:       errType,
		}
	case string:
		return &Err{
			Message:    v,
			StackTrace: string(debug.Stack()),
			Type:       errType,
		}
	case []byte:
		return &Err{
			Message:    string(v),
			StackTrace: string(debug.Stack()),
			Type:       errType,
		}
	default:
		panic("trying to create an Err with an unsupported type")
	}
}

var (
	ErrValidation = newErr("validation error", ErrTypeValidation)
)

var _ error = (*Err)(nil)
