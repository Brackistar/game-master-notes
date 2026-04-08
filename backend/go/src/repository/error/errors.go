package error

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound   = errors.New("repository: not found")
	ErrConflict   = errors.New("repository: conflict")
	ErrValidation = errors.New("repository: validation")
	ErrUnknown    = errors.New("repository: unknown")
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
