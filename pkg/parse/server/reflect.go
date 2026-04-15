package server

import (
	"reflect"
	"runtime"
	"slices"
	// e "github.com/simplicity-load/apispec/pkg/errors"
)

var intTypes = []reflect.Kind{
	reflect.Int,
	reflect.Int64,
	reflect.Int32,
	reflect.Int16,
	reflect.Int8,
}
var uintTypes = []reflect.Kind{
	reflect.Uint,
	reflect.Uint64,
	reflect.Uint32,
	reflect.Uint16,
	reflect.Uint8,
}
var primitiveTypes = slices.Concat(intTypes, uintTypes, []reflect.Kind{
	reflect.String,
	reflect.Bool,
})
var allowedFieldTypes = slices.Concat(primitiveTypes, []reflect.Kind{
	reflect.Struct,
	reflect.Array,
	reflect.Slice,
	reflect.Map,
})
var allowedMapKeyTypes = slices.Concat(intTypes, []reflect.Kind{
	reflect.String,
})

type qualifiedFnName string

func getFnName(fn any) (qualifiedFnName, error) {
	T := reflect.TypeOf(fn)
	if T.Kind() != reflect.Func {
		return "", FatalInvalidParam(T, reflect.Func)
	}

	ptr := reflect.ValueOf(fn).Pointer()
	rmFn := runtime.FuncForPC(ptr)
	return qualifiedFnName(rmFn.Name()), nil
}

func extractFieldType(s reflect.Type) (reflect.Type, error) { return extractFieldTypeIter(s, 0) }
func extractFieldTypeIter(s reflect.Type, indirectionLevel int) (reflect.Type, error) {
	if indirectionLevel == 2 {
		return nil, ErrSinglePointerRequired
	}

	K := s.Kind()
	switch {
	case K == reflect.Pointer:
		return extractFieldTypeIter(s.Elem(), indirectionLevel+1)
	case K == reflect.Float32, K == reflect.Float64:
		return nil, ErrNoFloat
	case slices.Contains(allowedFieldTypes, K):
		return s, nil
	default:
		return s, nil
		// TODO: extract sanity checking to post parsing
		// return nil, e.ErrBadValueFromList("type", K, allowedFieldTypes)
	}
}
