package http

import (
	"reflect"

	"github.com/simplicity-load/apispec/pkg/http"
)

type PathType string

const (
	PathROOT   PathType = "ROOT"
	PathSTATIC PathType = "STATIC"
	PathPARAM  PathType = "PARAM"
)

var ValidPathTypes = []PathType{
	PathROOT,
	PathSTATIC,
	PathPARAM,
}

type PathString struct {
	Name string
	Type PathType
}

type PathStrings []*PathString

type Endpoint struct {
	Method        http.Method   `json:",omitempty"`
	Path          []*PathString `json:",omitempty"`
	QueryParams   []string      `json:",omitempty"`
	UrlParams     []string      `json:",omitempty"`
	Authorization []string      `json:",omitempty"`
	Body          *Data         `json:",omitempty"`
	Response      *Data         `json:",omitempty"`
	Handler       *Handler      `json:",omitempty"`
}

type Handler struct {
	Name     string
	Import   string
	Reciever *Reciever
}

type Reciever struct {
	Name    string
	Pointer bool
}

type SerializationType string

const (
	SerializationJSON  SerializationType = "JSON"
	SerializationQUERY SerializationType = "QUERY"
	SerializationPATH  SerializationType = "PATH"
)

var ValidSerializationTypes = []SerializationType{
	SerializationJSON,
	SerializationQUERY,
	SerializationPATH,
}

type Serialization struct {
	Name string
	Type SerializationType
}

type StructField struct {
	Name          string         `json:",omitempty"`
	Type          reflect.Kind   `json:",omitempty"`
	Serialization *Serialization `json:",omitempty"`
	Validation    []string       `json:",omitempty"`
	SubFields     []*StructField `json:",omitempty"`
}

// Detailed information on an [Endpoint]'s request body or response
type Data struct {
	Name   string
	Import string
	Fields []*StructField
}

type Middleware = Handler

type Middlewares []*Middleware

type Path struct {
	PathString
	Endpoints  Endpoints   `json:",omitempty"`
	Middleware Middlewares `json:",omitempty"`
	SubPath    []*Path     `json:",omitempty"`
}

type Endpoints []*Endpoint

type Representation struct {
	Routes         *Path
	ErrorStatusMap map[error]int
}

// const a = `
// 	paths: [
// 		{
// 			path: "user",
// 			isParam: true,
// 			endpoints: [
// 				{
// 					method: GET,
// 					path: "user/blah/:name",
// 					pathparams: ["user", ":name"],
// 					queryparams: ["x","y","z"],
// 					authz: ["can_do_this", "can_do_that"],
// 					body: {
// 						name: "entity.GetUserByNameResponse",
// 						fields: {
// 							"blah": {
// 								validation: "required,blah,blah,blah",
// 								type: "query" | path | json // possibly file etc
// 								subfield:
// 							}
// 						}
// 					},
// 					res: "entity.GetUserByNameResponse",
// 					handler: "GetUserByName",
// 				}
// 			]
// 		}
// 	]
// `
