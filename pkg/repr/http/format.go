package http

import (
	"bytes"

	e "github.com/simplicity-load/apispec/pkg/errors"
)

func PathTypeToURL(path *PathString) (string, error) {
	url, ok := map[PathType]string{
		PathROOT:   "",
		PathPARAM:  ":" + path.Name,
		PathSTATIC: path.Name,
	}[path.Type]
	if !ok {
		return "", e.ErrBadValueFromList("path type", path.Type, ValidPathTypes)
	}
	return url, nil
}

func PathToURL(paths PathStrings) (string, error) {
	url := bytes.Buffer{}
	for path := range paths.NoRootPaths() {
		url.WriteRune('/')

		s, err := PathTypeToURL(path)
		if err != nil {
			return "", err
		}
		url.WriteString(s)
	}
	return url.String(), nil
}
