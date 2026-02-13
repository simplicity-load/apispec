package server

import (
	"context"
	"errors"
	"iter"
	"reflect"
	"slices"
	"strings"

	e "github.com/simplicity-load/apispec/pkg/errors"
	"github.com/simplicity-load/apispec/pkg/http"
	repr "github.com/simplicity-load/apispec/pkg/repr/http"
)

func ParsePaths(config *http.Path) (*repr.Path, error) {
	return traversePathsIter(config, repr.PathStrings{}, repr.Middlewares{})
}

func traversePathsIter(route *http.Path, paths []*repr.PathString, parentMiddleware repr.Middlewares) (*repr.Path, error) {
	ps := parsePathString(route)
	pathStrings := make([]*repr.PathString, 0, len(paths)+1)
	pathStrings = append(pathStrings, paths...)
	pathStrings = append(pathStrings, ps)

	endpoints, err := parseEndpoints(route.Endpoints, pathStrings)
	if err != nil {
		return nil, e.ErrFailedAction("parse endpoint", err)
	}

	middleware, err := parseMiddleware(route.Middleware)
	if err != nil {
		return nil, e.ErrFailedAction("parse middleware", err)
	}

	middlewareAcc := make(repr.Middlewares, 0, len(parentMiddleware)+len(middleware))
	middlewareAcc = append(middlewareAcc, parentMiddleware...)
	middlewareAcc = append(middlewareAcc, middleware...)

	subPaths := make([]*repr.Path, 0)
	for _, p := range route.SubPaths {
		path, err := traversePathsIter(p, pathStrings, middlewareAcc)
		if err != nil {
			return nil, err
		}
		subPaths = append(subPaths, path)
	}

	return &repr.Path{
		PathString: *ps,
		Endpoints:  endpoints,
		Middleware: middlewareAcc,
		SubPath:    subPaths,
	}, nil
}

func parseEndpoints(httpEndpoints http.Endpoints, paths []*repr.PathString) ([]*repr.Endpoint, error) {
	endpoints := make([]*repr.Endpoint, 0, len(httpEndpoints))

	for method, ep := range httpEndpoints {
		endpoint, err := parseHandler(ep, method, paths)
		if err != nil {
			return nil, ErrFnSignature(ep.Handler, err)
		}
		endpoints = append(endpoints, endpoint)
	}
	return endpoints, nil
}

func parseMiddleware(middlewareFns []any) (repr.Middlewares, error) {
	middleware := make(repr.Middlewares, 0, len(middlewareFns))
	for _, fn := range middlewareFns {
		handler, err := parseFnIdents(fn)
		if err != nil {
			return nil, e.ErrFailedAction("parse middleware function", err)
		}
		middleware = append(middleware, (*repr.Middleware)(handler))
	}
	return middleware, nil
}

func parseFnIdents(fn any) (*repr.Handler, error) {
	T := reflect.TypeOf(fn)
	if T.Kind() != reflect.Func {
		return nil, FatalInvalidParam(T, reflect.Func)
	}

	qualifiedName, err := getFnName(fn)
	if err != nil {
		return nil, e.ErrFailedAction("parse function name", err)
	}

	idents, err := parseIdentsFromQualifiedName(qualifiedName)
	if err != nil {
		return nil, e.ErrFailedAction("parse function identifiers from qualified name", err)
	}
	return idents, nil
}

var ctxInterface = reflect.TypeOf((*context.Context)(nil)).Elem()
var errInterface = reflect.TypeOf((*error)(nil)).Elem()

func parseHandler(ep http.Endpoint, method http.Method, paths []*repr.PathString) (*repr.Endpoint, error) {
	fn := reflect.TypeOf(ep.Handler)
	if fn.Kind() != reflect.Func {
		return nil, e.ErrBadType(fn, "function")
	}

	handler, err := parseFnIdents(ep.Handler)
	if err != nil {
		return nil, e.ErrFailedAction("parse function identifiers", err)
	}

	reqType, resType, err := parseFnSignature(fn)
	if err != nil {
		return nil, e.ErrFailedAction("parse function signature", err)
	}

	body, err := parseData(reqType)
	if err != nil {
		return nil, e.ErrFailedActionWithItem("parse body", reqType.Name(), err)
	}

	response, err := parseData(resType)
	if err != nil {
		return nil, e.ErrFailedActionWithItem("parse response", reqType.Name(), err)
	}

	// Parse endpoint-level middleware
	epMiddleware, err := parseMiddleware(ep.Middleware)
	if err != nil {
		return nil, e.ErrFailedAction("parse endpoint middleware", err)
	}

	return &repr.Endpoint{
		Method: method,
		Path:   paths,
		// QueryParams:   queryparams,
		Authorization: ep.Authz,
		Description:   ep.Description,
		Body:          body,
		Response:      response,
		Handler:       handler,
		Middleware:    epMiddleware,
	}, nil
}

func parseFnSignature(fn reflect.Type) (reflect.Type, reflect.Type, error) {
	const fnNumIn = 2
	if fn.NumIn() != fnNumIn {
		return nil, nil,
			e.ErrBadValue("parameter number", fn.NumIn(), fnNumIn)
	}
	const fnNumOut = 2
	if fn.NumOut() != fnNumOut {
		return nil, nil,
			e.ErrBadValue("result number", fn.NumOut(), fnNumOut)
	}

	ctxType := fn.In(0)
	reqPtrType := fn.In(1)
	resPtrType := fn.Out(0)
	errType := fn.Out(1)
	if !ctxType.Implements(ctxInterface) {
		return nil, nil, e.ErrBadType(ctxType, "context.Context")
	}

	if reqPtrType.Kind() != reflect.Pointer {
		return nil, nil, e.ErrBadType(reqPtrType, "*struct{...}")
	}
	reqType := reqPtrType.Elem()
	if reqType.Kind() != reflect.Struct {
		return nil, nil, e.ErrBadType(reqPtrType, "*struct{...}")
	}

	if resPtrType.Kind() != reflect.Pointer {
		return nil, nil, e.ErrBadType(resPtrType, "*struct{...}")
	}
	resType := resPtrType.Elem()
	if resType.Kind() != reflect.Struct {
		return nil, nil, e.ErrBadType(resPtrType, "*struct{...}")
	}

	if !errType.Implements(errInterface) {
		return nil, nil, e.ErrBadType(errType, "error")
	}
	return reqType, resType, nil
}

func parseData(s reflect.Type) (*repr.Data, error) {
	if s.Kind() != reflect.Struct {
		return nil, FatalInvalidParam(s, reflect.Struct)
	}

	body, err := parseBodyField(s)
	if err != nil {
		return nil, err
	}

	return &repr.Data{
		Name:   body.Name,
		Import: s.PkgPath(),
		Fields: body.SubFields,
	}, nil
}

func structFieldIter(s reflect.Type) iter.Seq[reflect.StructField] {
	return func(yield func(x reflect.StructField) bool) {
		if s.Kind() != reflect.Struct {
			return
		}

		for i := range s.NumField() {
			sf := s.Field(i)
			if isAnonStruct(sf) {
				for anonSf := range structFieldIter(sf.Type) {
					if !yield(anonSf) {
						return
					}
				}
				continue
			}
			if !yield(sf) {
				return
			}
		}
	}
}

func parseBodyField(s reflect.Type) (*repr.StructField, error) {
	T, err := extractFieldType(s)
	if err != nil {
		return nil, err
	}
	K := T.Kind()

	if slices.Contains(primitiveTypes, K) {
		return parsePrimitive(T)
	}

	switch K {
	case reflect.Array, reflect.Slice:
		return parseArraySlice(T)
	case reflect.Struct:
		return parseStruct(T)
	case reflect.Map:
		return parseMap(T)
	default:
		// Bad types are caught on type extraction
		return nil, ErrFatalUnreachable
	}
}

func parseArraySlice(s reflect.Type) (*repr.StructField, error) {
	bf, err := parseBodyField(s.Elem())
	if err != nil {
		return nil, err
	}
	return &repr.StructField{
		Type:      reflect.Array,
		SubFields: []*repr.StructField{bf},
	}, nil
}

func parseStruct(s reflect.Type) (*repr.StructField, error) {
	bfs := make([]*repr.StructField, 0, s.NumField())
	for f := range structFieldIter(s) {
		serialization, validation, err := parseFieldTag(f)
		if err != nil {
			return nil, e.ErrFailedActionWithItem("parse field tag", f.Name, err)
		}

		bf, err := parseBodyField(f.Type)
		if err != nil {
			return nil, e.ErrFailedActionWithItem("parse body field", f.Name, err)
		}
		bfs = append(bfs, &repr.StructField{
			Name:          f.Name,
			Serialization: serialization,
			Validation:    validation,
			Type:          bf.Type,
			SubFields:     bf.SubFields,
		})
	}
	return &repr.StructField{
		Name:      s.Name(),
		Type:      reflect.Struct,
		SubFields: bfs,
	}, nil
}

func parsePrimitive(s reflect.Type) (*repr.StructField, error) {
	return &repr.StructField{
		Type: s.Kind(),
	}, nil
}

func parseMap(s reflect.Type) (*repr.StructField, error) {
	T := s.Key()
	K := T.Kind()
	if !slices.Contains(allowedMapKeyTypes, T.Kind()) {
		return nil, e.ErrBadValueFromList("map key type", K, allowedMapKeyTypes)
	}

	v := s.Elem()
	vF, err := parseBodyField(v)
	if err != nil {
		return nil, err
	}
	return &repr.StructField{
		Type: reflect.Map,
		SubFields: []*repr.StructField{
			{
				Name: "key",
				Type: T.Kind(),
			},
			vF,
		},
	}, nil
}

func parseFieldTag(s reflect.StructField) (
	serialization *repr.Serialization,
	validation []string,
	err error,
) {
	if isAnonStruct(s) {
		return nil, nil, ErrFatalStructIsAnon
	}

	serialization, err = parseSerialization(s)
	if err != nil {
		return nil, nil,
			e.ErrFailedAction("parse serialization", err)
	}

	validation, err = parseValidation(s)
	if err != nil {
		return nil, nil, e.ErrFailedAction("parse validation", err)
	}
	return serialization, validation, nil
}

func isAnonStruct(s reflect.StructField) bool {
	return s.Type.Kind() == reflect.Struct && s.Anonymous
}

func parseSerialization(s reflect.StructField) (*repr.Serialization, error) {
	tag := s.Tag
	as, err := parseApiSpecTag(tag)
	if err != nil &&
		// ignore no value, check for json tag aswell
		!errors.Is(err, e.ErrNoValue) {
		return nil, e.ErrFailedActionWithItemWanted("parse tag", "as", "<serialization_name>,(path|query|header)", err)
	}
	if as != nil {
		return as, nil
	}

	json, err := parseJsonTag(tag)
	if err != nil {
		return nil, e.ErrFailedActionWithItemWanted("parse tag", "json", "<serialization_name>,...", err)
	}
	return json, nil
}

func parseApiSpecTag(t reflect.StructTag) (*repr.Serialization, error) {
	tag := t.Get("as")
	tagParts := strings.Split(tag, ",")
	if tag == "" {
		return nil, e.ErrNoValue
	}

	if len(tagParts) != 2 {
		return nil, e.ErrBadValue("number of tag values", len(tagParts), 2)
	}

	name := tagParts[0]
	if !isAllLowerA_Z(name) {
		return nil, ErrBadOnlyLowerAndDashFormatting(name)
	}
	serType := strings.ToUpper(tagParts[1])

	typedSerType := repr.SerializationType(serType)
	if !slices.Contains(repr.ApiSpecSerializationTypes, typedSerType) {
		return nil, e.ErrBadValueFromList(
			"serialization type",
			typedSerType,
			repr.ApiSpecSerializationTypes)
	}

	return &repr.Serialization{
		Name: name,
		Type: typedSerType,
	}, nil

}

func parseJsonTag(t reflect.StructTag) (*repr.Serialization, error) {
	tag := t.Get("json")
	if tag == "" {
		return nil, e.ErrBadValue("tag value", tag, "serialization_name")
	}
	tagParts := strings.Split(tag, ",")
	name := tagParts[0]
	if !isAllLowerA_Z(name) {
		return nil, ErrBadOnlyLowerAndDashFormatting(name)
	}
	return &repr.Serialization{
		Name: name,
		Type: repr.SerializationJSON,
	}, nil

}

func parseValidation(s reflect.StructField) (
	validation []string, err error,
) {
	val := s.Tag.Get("validate")
	if val == "" {
		return []string{}, nil // TODO(fati-kappe): fmt.Errorf("no validation tag found")
	}
	// TODO enforce only a subset of validation tags
	// or verify correct validation tags
	return strings.Split(val, ","), nil
}

func ParsePathString(path *http.Path) *repr.PathString {
	var t repr.PathType
	switch path.Type {
	case http.PathRoot:
		t = repr.PathROOT
	case http.PathStatic:
		t = repr.PathSTATIC
	case http.PathParam:
		t = repr.PathPARAM
	}
	return &repr.PathString{Name: path.Name, Type: t}
}

func parsePathString(path *http.Path) *repr.PathString {
	return ParsePathString(path)
}
