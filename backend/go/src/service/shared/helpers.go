package shared

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	repoerrors "github.com/Brackistar/game-master-notes/backend/go/src/repository/error"
	serviceerrors "github.com/Brackistar/game-master-notes/backend/go/src/service/error"
)

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}

func PanicIfNilDependency(serviceName, dependencyName string, dep any) {
	if isNil(dep) {
		panic(fmt.Sprintf(serviceerrors.SERVDEPNILMESSAGE, serviceName, dependencyName))
	}
}

func NormalizeSpaces(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func MapRepositoryError(err error, op, entity string) error {
	switch {
	case errors.Is(err, repoerrors.ErrNotFound):
		return serviceerrors.NewNotFound(op, entity)
	case errors.Is(err, repoerrors.ErrConflict):
		return serviceerrors.NewConflict(op, entity)
	case errors.Is(err, repoerrors.ErrValidation):
		return serviceerrors.WrapValidation(op, entity, err)
	case errors.Is(err, repoerrors.ErrUnknown):
		return serviceerrors.WrapUnknown(op, entity, err)
	default:
		return serviceerrors.WrapUnknown(op, entity, err)
	}
}

func isNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}
