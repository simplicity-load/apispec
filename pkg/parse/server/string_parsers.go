package server

import (
	"strings"

	e "github.com/simplicity-load/apispec/pkg/errors"
	repr "github.com/simplicity-load/apispec/pkg/repr/http"
)

func isFnAnonymousByQualifiedName(name qualifiedFnName) bool {
	return !strings.ContainsRune(string(name), '/')
}

func parseIdentsFromQualifiedName(qName qualifiedFnName) (*repr.Handler, error) {
	if isFnAnonymousByQualifiedName(qName) {
		return nil, ErrFnIsAnon
	}
	name := string(qName)

	pathIdx := strings.LastIndex(name, "/") // example.com/mypkg.User.GetName
	dottedIdentifiers := name[pathIdx:]     // mypkg.User.GetName

	pkg, fnName, reciever, err := parseDottedIdents(dottedIdentifiers)
	if err != nil {
		return nil, e.ErrFailedAction("parse dotted idents", err)
	}

	pkgPath := name[:pathIdx+len(pkg)]
	return &repr.Handler{
		Name:     fnName,
		Import:   pkgPath,
		Reciever: reciever,
	}, nil
}

func parseDottedIdents(dotIdents string) (
	pkgName string,
	fnName string,
	method *repr.Reciever,
	err error,
) {
	idents := strings.Split(dotIdents, ".")
	switch len(idents) {
	case 2: // mypkg.GetName
		return idents[0], idents[1], nil, nil
	case 3: // mypkg.User.GetName-fm
		name := strings.Split(idents[2], "-")[0] // GetUserByName-fm
		return idents[0], name, parseMethod(idents[1]), nil
	default:
		return "", "", nil, ErrFatalInvalidFnName
	}
}

func parseMethod(reciever string) *repr.Reciever {
	lastIdx := len(reciever) - 1
	if reciever[0] != '(' &&
		reciever[lastIdx] != ')' { // (*Service)
		return &repr.Reciever{
			Name:    reciever,
			Pointer: false,
		}
	}
	return &repr.Reciever{
		Name:    reciever[2:lastIdx],
		Pointer: true,
	}
}
