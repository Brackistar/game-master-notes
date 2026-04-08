package error

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound   = errors.New("service: not found")
	ErrConflict   = errors.New("service: conflict")
	ErrValidation = errors.New("service: validation")
	ErrUnknown    = errors.New("service: unknown")
)

func NewNotFound(op, entity string) error {
	return fmt.Errorf("%w: op=%s entity=%s", ErrNotFound, op, entity)
}

func NewConflict(op, entity string) error {
	return fmt.Errorf("%w: op=%s entity=%s", ErrConflict, op, entity)
}

func WrapValidation(op, entity string, cause error) error {
	return fmt.Errorf("%w: op=%s entity=%s cause=%v", ErrValidation, op, entity, cause)
}

func WrapUnknown(op, entity string, cause error) error {
	return fmt.Errorf("%w: op=%s entity=%s cause=%v", ErrUnknown, op, entity, cause)
}

