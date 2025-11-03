package server

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
)

// *-------------*
// | USER ERRORS |
// *-------------*

func ErrFnSignature(got any, err ...error) error {
	t := reflect.TypeOf(got)
	signature := bytes.NewBufferString(fmt.Sprintf(
		`Handlers must have the following signature:
    expected: func(context.Context, *struct{...}) (*struct{...}, error)
         got: %s`,
		t,
	))
	for _, e := range err {
		fmt.Fprintf(signature,
			"\n       error: %s",
			e,
		)
	}
	return errors.New(signature.String())
}

func ErrBodyTypeError(got any, err error) error {
	// TODO
	return nil
}

func ErrBadOnlyLowerFormatting(got string) error {
	return errBadFormatting(got, "lowercase a to z")
}
func ErrBadOnlyLowerAndDashFormatting(got string) error {
	return errBadFormatting(got, "lowercase a to z and '_'")
}

func errBadFormatting(got, want string) error {
	return fmt.Errorf(`accepted characters: %q, got: %q`, want, got)
}

var (
	ErrFnIsAnon              = errors.New("function must not be anonymous")
	ErrSinglePointerRequired = errors.New("single pointer required")
	ErrNoFloat               = errors.New("floats aren't allowed")
)

// *-----------------*
// | INTERNAL ERRORS |
// *-----------------*
func FatalInvalidParam(p reflect.Type, want reflect.Kind) error {
	return fmt.Errorf(
		`fatal error: invalid parameter passed: %s, wanted: %s`,
		p.Kind(),
		want,
	)
}

var (
	ErrFatalInvalidFnName = errors.New("fatal error: function name is invalid")
	ErrFatalStructIsAnon  = errors.New("fatal error: struct is anonymous")
	ErrFatalUnreachable   = errors.New("fatal error: unreachable")
)
