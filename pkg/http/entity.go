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
	Handler     any
	Description string
	Authz       []string
	Middleware  []any
}

type Endpoints map[Method]Endpoint

type optionType int

const (
	optAuthz optionType = iota
	optMiddleware
)

type EndpointOpt struct {
	typ    optionType
	authz  []string
	middle []any
}

func (o EndpointOpt) getOptionType() optionType { return o.typ }

func Authz(values ...string) EndpointOpt {
	return EndpointOpt{typ: optAuthz, authz: values}
}

func Middleware(values ...any) EndpointOpt {
	return EndpointOpt{typ: optMiddleware, middle: values}
}

type PathType int

const (
	PathRoot PathType = iota
	PathStatic
	PathParam
)

type Path struct {
	Name       string
	Type       PathType
	Endpoints  map[Method]Endpoint
	Middleware []any
	SubPaths   []*Path
}

func NewAPI() *Path {
	return &Path{Type: PathRoot, Endpoints: make(map[Method]Endpoint)}
}

func (p *Path) Static(name string) *Path {
	child := &Path{Name: name, Type: PathStatic, Endpoints: make(map[Method]Endpoint)}
	p.SubPaths = append(p.SubPaths, child)
	return child
}

func (p *Path) Param(name string) *Path {
	child := &Path{Name: name, Type: PathParam, Endpoints: make(map[Method]Endpoint)}
	p.SubPaths = append(p.SubPaths, child)
	return child
}

func (p *Path) Use(middleware ...any) {
	p.Middleware = append(p.Middleware, middleware...)
}

func (p *Path) addEndpoint(method Method, handler any, desc string, opts []EndpointOpt) {
	ep := Endpoint{Handler: handler, Description: desc}
	for _, opt := range opts {
		switch opt.getOptionType() {
		case optAuthz:
			ep.Authz = opt.authz
		case optMiddleware:
			ep.Middleware = opt.middle
		}
	}
	p.Endpoints[method] = ep
}

func (p *Path) Get(handler any, desc string, opts ...EndpointOpt) {
	p.addEndpoint(GET, handler, desc, opts)
}

func (p *Path) Post(handler any, desc string, opts ...EndpointOpt) {
	p.addEndpoint(POST, handler, desc, opts)
}

func (p *Path) Put(handler any, desc string, opts ...EndpointOpt) {
	p.addEndpoint(PUT, handler, desc, opts)
}

func (p *Path) Patch(handler any, desc string, opts ...EndpointOpt) {
	p.addEndpoint(PATCH, handler, desc, opts)
}

func (p *Path) Delete(handler any, desc string, opts ...EndpointOpt) {
	p.addEndpoint(DELETE, handler, desc, opts)
}

type HttpServer struct {
	ServerTemplate string
	ClientTemplate string
	Routes         *Path
	OutputFile     io.Writer
	ValidateUrl    string
}

type OpenAPIConfig struct {
	Routes     *Path
	OutputFile io.Writer
	Title      string
	Version    string
	ServerURL  string
}
