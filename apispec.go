package apispec

import (
	"errors"
	"fmt"

	generate "github.com/simplicity-load/apispec/pkg/gen"
	"github.com/simplicity-load/apispec/pkg/gen/openapi"
	"github.com/simplicity-load/apispec/pkg/http"
	"github.com/simplicity-load/apispec/pkg/parse/server"
	repr "github.com/simplicity-load/apispec/pkg/repr/http"
)

func ParseServer(root *http.Path) (*repr.Path, error) {
	paths, err := server.ParsePaths(root)
	if err != nil {
		return nil, fmt.Errorf("failed traversing paths: %w", err)
	}
	return paths, nil
}

func Generate(routes *repr.Path, config http.HttpServer) error {
	// for servers
	// get all info, such as pathparams and queryparams
	// start constructing endpoints based on templates
	// output them on specified folders
	//
	// for clients
	// get all info, such as pathparams and queryparams
	// start constructing callers based on templates
	// output them on specified folders
	err := generate.Generate(
		routes,
		config.OutputFile,
		config.ValidateUrl)
	if err != nil {
		return fmt.Errorf("failed generating: %w", err)
	}
	return nil
}

func GenerateOpenAPI(routes *repr.Path, config http.OpenAPIConfig) error {
	if config.Title == "" {
		return errors.New("missing title")
	}
	if config.Version == "" {
		return errors.New("missing version")
	}

	err := openapi.Generate(
		routes,
		config.OutputFile,
		config.Title,
		config.Version,
		config.ServerURL,
	)
	if err != nil {
		return fmt.Errorf("failed generating OpenAPI spec: %w", err)
	}

	return nil
}
