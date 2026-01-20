package openapi

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	repr "github.com/simplicity-load/apispec/pkg/repr/http"
)

// Generate creates an OpenAPI v3.1 specification from the parsed routes
func Generate(routes *repr.Path, output io.Writer, title, version, serverURL string) error {
	spec := &OpenAPI{
		OpenAPI: "3.1.0",
		Info: Info{
			Title:   title,
			Version: version,
		},
		Paths: make(map[string]PathItem),
	}

	if serverURL != "" {
		spec.Servers = []Server{
			{URL: serverURL},
		}
	}

	// Convert routes to OpenAPI paths
	if err := convertPaths(routes, spec.Paths, ""); err != nil {
		return fmt.Errorf("failed to convert paths: %w", err)
	}

	// Write JSON output
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(spec); err != nil {
		return fmt.Errorf("failed to encode OpenAPI spec: %w", err)
	}

	return nil
}

// convertPaths recursively converts repr.Path to OpenAPI paths
func convertPaths(path *repr.Path, paths map[string]PathItem, parentPath string) error {
	// Build the path string
	pathStr := buildPathString(path)

	fullPath := parentPath
	if pathStr != "/" {
		if fullPath == "/" {
			fullPath = pathStr
		} else {
			fullPath = fullPath + pathStr
		}
	} else if fullPath == "" {
		fullPath = "/"
	}

	// Convert endpoints for this path
	if len(path.Endpoints) > 0 {
		pathItem := paths[fullPath]
		for _, endpoint := range path.Endpoints {
			operation := convertEndpoint(endpoint)
			setOperation(&pathItem, string(endpoint.Method), operation)
		}
		paths[fullPath] = pathItem
	}

	// Process subpaths recursively
	for _, subPath := range path.SubPath {
		if err := convertPaths(subPath, paths, fullPath); err != nil {
			return err
		}
	}

	return nil
}

// buildPathString constructs the full path string from a repr.Path
func buildPathString(path *repr.Path) string {
	var parts []string

	// The path itself contains the full path hierarchy
	if path.Type == repr.PathROOT {
		return "/"
	}

	if path.Type == repr.PathSTATIC {
		parts = append(parts, path.Name)
	} else if path.Type == repr.PathPARAM {
		parts = append(parts, "{"+path.Name+"}")
	}

	if len(parts) == 0 {
		return "/"
	}

	return "/" + strings.Join(parts, "/")
}

// convertEndpoint converts a repr.Endpoint to an OpenAPI Operation
func convertEndpoint(endpoint *repr.Endpoint) *Operation {
	operation := &Operation{
		Summary:     endpoint.Handler.Name,
		OperationID: endpoint.Handler.Name,
		Responses:   make(map[string]Response),
	}

	// Add parameters (path and query)
	var params []Parameter
	for _, field := range endpoint.Body.Fields {
		if field.Serialization != nil {
			switch field.Serialization.Type {
			case repr.SerializationPATH:
				params = append(params, Parameter{
					Name:     field.Serialization.Name,
					In:       "path",
					Required: true,
					Schema:   convertFieldToSchema(field),
				})
			case repr.SerializationQUERY:
				params = append(params, Parameter{
					Name:     field.Serialization.Name,
					In:       "query",
					Required: isRequired(field.Validation),
					Schema:   convertFieldToSchema(field),
				})
			}
		}
	}
	if len(params) > 0 {
		operation.Parameters = params
	}

	// Add request body for non-GET methods
	if endpoint.Method != "GET" {
		bodySchema := convertDataToSchema(endpoint.Body)
		if bodySchema != nil && len(bodySchema.Properties) > 0 {
			operation.RequestBody = &RequestBody{
				Required: true,
				Content: map[string]MediaType{
					"application/json": {
						Schema: bodySchema,
					},
				},
			}
		}
	}

	// Add response
	responseSchema := convertDataToSchema(endpoint.Response)
	operation.Responses["200"] = Response{
		Description: "Successful response",
		Content: map[string]MediaType{
			"application/json": {
				Schema: responseSchema,
			},
		},
	}

	return operation
}

// convertDataToSchema converts repr.Data to an inline OpenAPI Schema
func convertDataToSchema(data *repr.Data) *Schema {
	if data == nil {
		return &Schema{Type: "object"}
	}

	schema := &Schema{
		Type:       "object",
		Properties: make(map[string]*Schema),
	}

	var required []string
	for _, field := range data.Fields {
		// Skip path and query params as they're handled separately
		if field.Serialization != nil &&
			(field.Serialization.Type == repr.SerializationPATH ||
				field.Serialization.Type == repr.SerializationQUERY) {
			continue
		}

		fieldSchema := convertFieldToSchema(field)
		fieldName := field.Name
		if field.Serialization != nil {
			fieldName = field.Serialization.Name
		}

		schema.Properties[fieldName] = fieldSchema

		if isRequired(field.Validation) {
			required = append(required, fieldName)
		}
	}

	if len(required) > 0 {
		schema.Required = required
	}

	return schema
}

// convertFieldToSchema converts a repr.StructField to an OpenAPI Schema
func convertFieldToSchema(field *repr.StructField) *Schema {
	schema := &Schema{}

	switch field.Type {
	case reflect.String:
		schema.Type = "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema.Type = "integer"
	case reflect.Float32, reflect.Float64:
		schema.Type = "number"
	case reflect.Bool:
		schema.Type = "boolean"
	case reflect.Array, reflect.Slice:
		schema.Type = "array"
		if len(field.SubFields) > 0 {
			schema.Items = convertFieldToSchema(field.SubFields[0])
		} else {
			schema.Items = &Schema{Type: "string"}
		}
	case reflect.Struct:
		schema.Type = "object"
		schema.Properties = make(map[string]*Schema)
		var required []string
		for _, subField := range field.SubFields {
			subFieldName := subField.Name
			if subField.Serialization != nil {
				subFieldName = subField.Serialization.Name
			}
			schema.Properties[subFieldName] = convertFieldToSchema(subField)
			if isRequired(subField.Validation) {
				required = append(required, subFieldName)
			}
		}
		if len(required) > 0 {
			schema.Required = required
		}
	case reflect.Map:
		schema.Type = "object"
		// For maps, we use additionalProperties pattern
		if len(field.SubFields) > 1 {
			// SubFields[0] is the key, SubFields[1] is the value
			// OpenAPI doesn't directly support typed keys, so we just type the values
			schema.Properties = make(map[string]*Schema)
		}
	default:
		schema.Type = "string"
	}

	return schema
}

// setOperation sets the operation on the path item based on HTTP method
func setOperation(pathItem *PathItem, method string, operation *Operation) {
	switch strings.ToUpper(method) {
	case "GET":
		pathItem.Get = operation
	case "POST":
		pathItem.Post = operation
	case "PUT":
		pathItem.Put = operation
	case "PATCH":
		pathItem.Patch = operation
	case "DELETE":
		pathItem.Delete = operation
	}
}

// isRequired checks if a field has "required" in its validation tags
func isRequired(validation []string) bool {
	for _, v := range validation {
		if v == "required" {
			return true
		}
	}
	return false
}
