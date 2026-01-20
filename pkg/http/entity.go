package http

import "io"

type Method string

const (
	POST   Method = "POST"
	GET    Method = "GET"
	PUT    Method = "PUT"
	PATCH  Method = "PATCH"
	DELETE Method = "DELETE"
)

type Endpoint struct {
	Handler any
	Authn   []string
}

type Endpoints map[Method]Endpoint

var ValidPaths = []Path{
	RootPath{},
	StaticPath{},
	ParamPath{},
}

type RootPath struct {
	Endpoints  Endpoints
	Middleware []any
	SubPaths   []Path
}

func (p RootPath) isPath()             {}
func (p RootPath) SubPath() []Path     { return p.SubPaths }
func (p RootPath) Endpoint() Endpoints { return p.Endpoints }

type StaticPath struct {
	Path       string
	Endpoints  Endpoints
	Middleware []any
	SubPaths   []Path
}

func (p StaticPath) isPath()             {}
func (p StaticPath) SubPath() []Path     { return p.SubPaths }
func (p StaticPath) Endpoint() Endpoints { return p.Endpoints }

type ParamPath struct {
	Path      string
	Endpoints Endpoints
	SubPaths  []Path
}

func (p ParamPath) isPath()             {}
func (p ParamPath) SubPath() []Path     { return p.SubPaths }
func (p ParamPath) Endpoint() Endpoints { return p.Endpoints }

type Path interface {
	isPath()
	SubPath() []Path
	Endpoint() Endpoints
}

type HttpServer struct {
	ServerTemplate string
	ClientTemplate string
	Routes         RootPath
	OutputFile     io.Writer
	ValidateUrl    string
}
