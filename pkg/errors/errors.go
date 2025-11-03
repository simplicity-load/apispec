package errors

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// *-------------*
// | USER ERRORS |
// *-------------*

func ErrBadValue[T any](target string, got T, want T) error {
	return fmt.Errorf(`incorrect %q: "%v", required: "%v"`,
		target,
		got,
		want)
}

func ErrBadType(got reflect.Type, want string) error {
	return fmt.Errorf(`incorrect type: %q, required: %q`,
		got,
		want)
}

func ErrBadValueFromList[T any](target string, got T, want []T) error {
	values := make([]string, 0, len(want))
	for _, v := range want {
		s := fmt.Sprintf("%v", v)
		values = append(values, s)
	}
	return fmt.Errorf(
		`invalid %q: "%v", allowed values are: %q`,
		target,
		got,
		strings.Join(values, ", "))
}

func ErrFailedAction(action string, err error) error {
	return fmt.Errorf(`failed %q, err: (%w)`, action, err)
}

func ErrFailedActionWithItem(action, actionItem string, err error) error {
	return fmt.Errorf(`failed %q, on %q, err: (%w)`, action, actionItem, err)
}

func ErrFailedActionWithItemWanted(action, actionItem, wanted string, err error) error {
	return fmt.Errorf(`failed %q, on %q, wanted: %q, err: (%w)`, action, actionItem, wanted, err)
}

var ErrNoValue = errors.New("missing value")

// *-----------------*
// | FATAL ERRORS    |
// *-----------------*

var (
	ErrFatalUnreachable = errors.New("fatal error: unreachable")
)
