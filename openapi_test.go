package apispec_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/simplicity-load/apispec"
	"github.com/simplicity-load/apispec/pkg/http"
)

// Define complex data structures for testing
type Address struct {
	Street string `json:"street"`
	City   string `json:"city"`
}

type User struct {
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	Emails  []string        `json:"emails"`
	Address Address         `json:"address"`
	Roles   map[string]bool `json:"roles"`
}

type CreateUserRequest struct {
	Name    string  `json:"name" validate:"required"`
	Email   string  `json:"email" validate:"required"`
	Address Address `json:"address"`
}

type UpdateUserRequest struct {
	ID      string  `json:"id" as:"id,path" validate:"required"`
	Name    string  `json:"name"`
	Address Address `json:"address"`
}

type GetUserRequest struct {
	ID string `json:"id" as:"id,path" validate:"required"`
}

type ListUsersRequest struct {
	Limit  int `json:"limit" as:"limit,query"`
	Offset int `json:"offset" as:"offset,query"`
}

type EmptyResponse struct{}

type ListUsersResponse struct {
	Users []User `json:"users"`
}

func TestGenerateOpenAPI_Comprehensive(t *testing.T) {
	// Define handlers
	createUserHandler := func(ctx context.Context, req *CreateUserRequest) (*User, error) { return nil, nil }
	updateUserHandler := func(ctx context.Context, req *UpdateUserRequest) (*User, error) { return nil, nil }
	getUserHandler := func(ctx context.Context, req *GetUserRequest) (*User, error) { return nil, nil }
	listUsersHandler := func(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error) { return nil, nil }
	deleteUserHandler := func(ctx context.Context, req *GetUserRequest) (*EmptyResponse, error) { return nil, nil }

	// Create route configuration
	// Expected paths:
	// POST /              (Create User - though usually /users, testing root path here)
	// GET /               (List Users)
	// GET /users/{id}     (Get User)
	// PUT /users/{id}     (Update User)
	// DELETE /users/{id}  (Delete User)
	routes := http.RootPath{
		Endpoints: http.Endpoints{
			http.POST: http.Endpoint{Handler: createUserHandler},
			http.GET:  http.Endpoint{Handler: listUsersHandler},
		},
		SubPaths: []http.Path{
			http.StaticPath{
				Path: "users",
				SubPaths: []http.Path{
					http.ParamPath{
						Path: "id",
						Endpoints: http.Endpoints{
							http.GET:    http.Endpoint{Handler: getUserHandler},
							http.PUT:    http.Endpoint{Handler: updateUserHandler},
							http.DELETE: http.Endpoint{Handler: deleteUserHandler},
						},
					},
				},
			},
		},
	}

	// Create output buffer
	var output bytes.Buffer

	// Create config
	config := http.OpenAPIConfig{
		Routes:     routes,
		OutputFile: &output,
		Title:      "Complex User API",
		Version:    "2.0.0",
		ServerURL:  "https://api.example.com",
	}

	// Generate OpenAPI spec
	err := apispec.GenerateOpenAPI(config)
	if err != nil {
		t.Fatalf("GenerateOpenAPI failed: %v", err)
	}

	// Parse the output to verify structure
	var result map[string]interface{}
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// 1. Verify Info and Server
	info := result["info"].(map[string]interface{})
	if info["title"] != "Complex User API" {
		t.Errorf("Title mismatch: expected 'Complex User API', got '%s'", info["title"])
	}
	servers := result["servers"].([]interface{})
	if len(servers) != 1 || servers[0].(map[string]interface{})["url"] != "https://api.example.com" {
		t.Errorf("Server URL mismatch")
	}

	paths := result["paths"].(map[string]interface{})

	// 2. Verify Root Path (/)
	rootPath, ok := paths["/"]
	if !ok {
		t.Fatalf("Root path '/' not found in spec")
	}
	rootMap := rootPath.(map[string]interface{})

	// Check GET / (List Users) - Query Params
	if _, ok := rootMap["get"]; !ok {
		t.Error("GET method missing on root path")
	} else {
		getOp := rootMap["get"].(map[string]interface{})
		params := getOp["parameters"].([]interface{})

		// Verify query parameters limit and offset
		foundLimit := false
		foundOffset := false
		for _, p := range params {
			param := p.(map[string]interface{})
			if param["name"] == "limit" && param["in"] == "query" {
				foundLimit = true
			}
			if param["name"] == "offset" && param["in"] == "query" {
				foundOffset = true
			}
		}
		if !foundLimit || !foundOffset {
			t.Error("Missing query parameters 'limit' or 'offset' on GET /")
		}
	}

	// Check POST / (Create User) - Request Body and Nested Structs
	if _, ok := rootMap["post"]; !ok {
		t.Error("POST method missing on root path")
	} else {
		postOp := rootMap["post"].(map[string]interface{})

		// Check request body
		if postOp["requestBody"] == nil {
			t.Error("Missing request body on POST /")
		} else {
			rb := postOp["requestBody"].(map[string]interface{})
			content := rb["content"].(map[string]interface{})
			jsonContent := content["application/json"].(map[string]interface{})
			schema := jsonContent["schema"].(map[string]interface{})
			props := schema["properties"].(map[string]interface{})

			// Verify flattened/nested structure
			if _, ok := props["name"]; !ok {
				t.Error("Missing 'name' field in CreateUserRequest schema")
			}
			if addr, ok := props["address"]; !ok {
				t.Error("Missing 'address' nested struct in CreateUserRequest schema")
			} else {
				addrProps := addr.(map[string]interface{})["properties"].(map[string]interface{})
				if _, ok := addrProps["city"]; !ok {
					t.Error("Missing 'city' field in nested Address schema")
				}
			}
		}
	}

	// 3. Verify Subpath /users/{id}
	// Note: Path construction logic: / + users + / + {id}
	idPathKey := "/users/{id}"
	idPath, ok := paths[idPathKey]
	if !ok {
		// Fallback check if path generation differs slightly
		found := false
		for k := range paths {
			if strings.Contains(k, "users") && strings.Contains(k, "{id}") {
				idPathKey = k
				idPath = paths[k]
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Path '/users/{id}' not found in spec. Available paths: %v", getKeys(paths))
		}
	}
	idMap := idPath.(map[string]interface{})

	// Check PUT /users/{id} - Path Param and Body
	if _, ok := idMap["put"]; !ok {
		t.Error("PUT method missing on " + idPathKey)
	} else {
		putOp := idMap["put"].(map[string]interface{})
		params := putOp["parameters"].([]interface{})

		// Verify Id path parameter
		foundId := false
		for _, p := range params {
			param := p.(map[string]interface{})
			if param["name"] == "id" && param["in"] == "path" && param["required"] == true {
				foundId = true
			}
		}
		if !foundId {
			t.Error("Missing path parameter 'id' on PUT " + idPathKey)
		}
	}

	// Check GET response schema (Wrapped ListUsersResponse)
	if getOp, ok := rootMap["get"].(map[string]interface{}); ok {
		responses := getOp["responses"].(map[string]interface{})
		okResp := responses["200"].(map[string]interface{})
		content := okResp["content"].(map[string]interface{})
		jsonContent := content["application/json"].(map[string]interface{})
		schema := jsonContent["schema"].(map[string]interface{})

		// Expect object type (ListUsersResponse struct)
		if schema["type"] != "object" {
			t.Errorf("Expected response type 'object' for GET /, got '%v'", schema["type"])
		}
		props := schema["properties"].(map[string]interface{})
		if _, ok := props["users"]; !ok {
			t.Error("Missing 'users' field in ListUsersResponse schema")
		}
		usersProp := props["users"].(map[string]interface{})
		if usersProp["type"] != "array" {
			t.Errorf("Expected 'users' property to be 'array', got '%v'", usersProp["type"])
		}
	}
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
