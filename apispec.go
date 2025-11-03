package apispec

import (
	"fmt"

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
	err = generate(repr.Representation{
		Routes: paths,
	}, config.OutputFile)
	if err != nil {
		return fmt.Errorf("failed generating: %w", err)
	}
	return nil
}
