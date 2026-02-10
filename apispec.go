package apispec

import (
	"fmt"

	generate "github.com/simplicity-load/apispec/pkg/gen"
	"github.com/simplicity-load/apispec/pkg/gen/openapi"
	"github.com/simplicity-load/apispec/pkg/http"
	"github.com/simplicity-load/apispec/pkg/parse/server"
	repr "github.com/simplicity-load/apispec/pkg/repr/http"
)

func Generate(config http.HttpServer) error {
	// for servers
	// get all info, such as pathparams and queryparams
	// start constructing endpoints based on templates
	// output them on specified folders
	//
	// for clients
	// get all info, such as pathparams and queryparams
	// start constructing callers based on templates
	// output them on specified folders
	paths, err := server.ParsePaths(config.Routes)
	if err != nil {
		return fmt.Errorf("failed traversing paths: %w", err)
	}
	err = generate.Generate(repr.Representation{
		Routes: paths,
	}, config.OutputFile, config.ValidateUrl)
	if err != nil {
		return fmt.Errorf("failed generating: %w", err)
	}
	return nil
}

func GenerateOpenAPI(config http.OpenAPIConfig) error {
	// Parse the routes to get structured representation
	paths, err := server.ParsePaths(config.Routes)
	if err != nil {
		return fmt.Errorf("failed parsing paths: %w", err)
	}

	// Set defaults for optional fields
	title := config.Title
	if title == "" {
		title = "API Specification"
	}
	version := config.Version
	if version == "" {
		version = "1.0.0"
	}

	// Generate OpenAPI specification
	err = openapi.Generate(paths, config.OutputFile, title, version, config.ServerURL)
	if err != nil {
		return fmt.Errorf("failed generating OpenAPI spec: %w", err)
	}

	return nil
}
