package apispec

import (
	"bytes"
	"fmt"
	"io"
	"iter"
	"slices"
	"strconv"
	"strings"
	"sync"
	"text/template"

	"github.com/simplicity-load/apispec/pkg/http"
	repr "github.com/simplicity-load/apispec/pkg/repr/http"
)

type set[T comparable] map[T]uint

func newSet[T comparable]() set[T] {
	return make(set[T])
}

func (s set[T]) get(x T) uint {
	return s[x]
}

func (s set[T]) add(x T) {
	if _, ok := s[x]; ok {
		return
	}
	s[x] = uint(len(s))
}

type importSet struct {
	s set[string]
}

func newImportSet() importSet {
	return importSet{s: newSet[string]()}
}

func (i importSet) ident(count uint) string {
	return intToIdent(count) + "i"
}

func (i importSet) get(imp string) string {
	count := i.s.get(imp)
	return i.ident(count)
}

func (i importSet) add(imp string) {
	i.s.add(imp)
}

type importer struct {
	Import string
	Ident  string
}

func (i importSet) all() iter.Seq[importer] {
	return func(yield func(importer) bool) {
		for k, v := range i.s {
			if !yield(importer{
				Import: k,
				Ident:  i.ident(v),
			}) {
				return
			}
		}
	}
}

func intToIdent(x uint) (ident string) {
	r := x % 26
	m := x / 26
	for range m {
		ident += "z"
	}
	return string(rune('a'+r)) + ident
}

type recieverSet struct {
	s set[string]
}

func newRecieverSet() recieverSet {
	return recieverSet{s: newSet[string]()}
}

func (r recieverSet) key(x repr.Reciever, importPath string) string {
	return fmt.Sprintf("%s\n%s\n%t", importPath, x.Name, x.Pointer)
}

func (r recieverSet) ident(count uint) string {
	return intToIdent(count) + "r"
}

func (r recieverSet) get(x repr.Reciever, importPath string) string {
	count := r.s.get(r.key(x, importPath))
	return r.ident(count)
}

func (r recieverSet) add(x repr.Reciever, importPath string) {
	r.s.add(r.key(x, importPath))
}

type reciever struct {
	repr.Reciever
	Import string
	Ident  string
}

func (r recieverSet) all() iter.Seq[reciever] {
	return func(yield func(reciever) bool) {
		for k, v := range r.s {
			str := strings.Split(k, "\n")
			if len(str) != 3 {
				panic(fmt.Sprintf("invalid key: %s", k))
			}
			b, err := strconv.ParseBool(str[2])
			if err != nil {
				panic(fmt.Sprintf("invalid key: %s, err: %s", k, err))
			}
			if !yield(reciever{
				Reciever: repr.Reciever{
					Name:    str[1],
					Pointer: b,
				},
				Import: str[0],
				Ident:  r.ident(v),
			}) {
				return
			}
		}
	}
}

func generate(
	representation repr.Representation,
	output io.Writer,
	validateUrl string,
) error {
	imports := newImportSet()
	recievers := newRecieverSet()
	endpoints, err := generateEndpoints(representation.Routes, imports, recievers)
	if err != nil {
		return err
	}

	t, err := template.New("").Funcs(template.FuncMap{
		"importIdent": func(imp string) string {
			return imports.get(imp)
		},
		"isPointer": func(isPointer bool) string {
			if isPointer {
				return "*"
			}
			return ""
		},
	}).Parse(registerTemplate)
	if err != nil {
		panic(fmt.Sprintf("template parsing failed, err: %s", err))
	}

	gen, err := templateToString(t, struct {
		Recievers      iter.Seq[reciever]
		Imports        iter.Seq[importer]
		Endpoints      []string
		ValidateImport string
	}{
		Recievers:      recievers.all(),
		Imports:        imports.all(),
		Endpoints:      endpoints,
		ValidateImport: validateUrl,
	})
	if err != nil {
		panic(err)
	}
	_, err = fmt.Fprintln(output, gen)
	if err != nil {
		panic(err)
	}
	return nil
}

func generateEndpoints(path *repr.Path, imports importSet, recievers recieverSet) (
	[]string, error,
) {
	return generateEndpointsIter(path, imports, recievers)
}

func generateEndpointsIter(path *repr.Path, imports importSet, recievers recieverSet) (
	[]string, error,
) {
	for imp := range path.Endpoints.Imports() {
		imports.add(imp)
	}
	for reciever, imp := range path.Endpoints.Recievers() {
		recievers.add(reciever, imp)
	}
	for imp := range path.Middleware.Imports() {
		imports.add(imp)
	}

	endpoints := make([]string, 0, len(path.Endpoints))
	templatedEndpoints := templatedEndpoints(path.Endpoints, path.Middleware, imports, recievers)
	endpoints = slices.AppendSeq(endpoints, templatedEndpoints)

	for _, subPath := range path.SubPath {
		lastEndpoints, err := generateEndpointsIter(subPath, imports, recievers)
		if err != nil {
			return nil, err
		}
		endpoints = append(endpoints, lastEndpoints...)
	}
	return endpoints, nil
}

var getEndpointTemplate = sync.OnceValue(func() *template.Template {
	t, err := template.New("").Funcs(template.FuncMap{
		"pathToString":      repr.PathToURL,
		"httpMethodToFiber": httpMethodToFiber,
		"generateParams":    generateParams,
		"formatMiddleware":  formatMiddleware,
	}).Parse(endpointTemplate)
	if err != nil {
		panic(fmt.Sprintf("template parsing failed, template: %s, err: %s", t.DefinedTemplates(), err))
	}
	return t
})

var getParamsTemplate = sync.OnceValue(func() *template.Template {
	t, err := template.New("").Parse(
		`body.{{ .Name }} = {{ .FunctionName }}("{{ .Serialization }}")`,
	)
	if err != nil {
		panic(fmt.Sprintf("template parsing failed, template: %s, err: %s", t.DefinedTemplates(), err))
	}
	return t
})

func templatedEndpoints(endpoints []*repr.Endpoint, middleware repr.Middlewares, imports importSet, recievers recieverSet) iter.Seq[string] {
	t := getEndpointTemplate()

	type templateData struct {
		*repr.Endpoint
		IsGet         bool
		ImportIdent   string
		RecieverIdent string
		Middleware    repr.Middlewares
		Imports       importSet
	}
	data := templateData{
		Middleware: middleware,
		Imports:    imports,
	}

	return func(yield func(x string) bool) {
		for _, endpoint := range endpoints {
			data.Endpoint = endpoint
			data.IsGet = endpoint.Method == http.GET
			data.ImportIdent = imports.get(endpoint.Body.Import)
			data.RecieverIdent = recievers.get(*endpoint.Handler.Reciever, endpoint.Handler.Import)

			s, err := templateToString(t, data)
			if err != nil {
				panic(err)
			}
			braced := "{" + endpoint.Handler.Name + "}"
			fmt.Printf("[✔] Endpoint generated %s\t\n", braced)
			if !yield(s) {
				return
			}
		}
	}

}

func httpMethodToFiber(method http.Method) string {
	x := string(method)
	return strings.ToUpper(x[:1]) + strings.ToLower(x[1:])
}

func formatMiddleware(middleware repr.Middlewares, imports importSet) iter.Seq[string] {
	return func(yield func(x string) bool) {
		for _, m := range middleware {
			var formatted string
			if m.Reciever != nil {
				formatted = fmt.Sprintf("%s.%s",
					imports.get(m.Import),
					m.Name)
			} else {
				formatted = fmt.Sprintf("%s.%s",
					imports.get(m.Import),
					m.Name)
			}
			if !yield(formatted) {
				return
			}
		}
	}
}

func serializationToFunctionName(t repr.SerializationType) (string, bool) {
	fnName, ok := map[repr.SerializationType]string{
		repr.SerializationPATH:  "c.Params",
		repr.SerializationQUERY: "c.Query",
	}[t]
	return fnName, ok
}

func toSimplifiedParams(body *repr.Data) iter.Seq[*param] {
	return func(yield func(x *param) bool) {
		for _, fields := range body.Fields {
			fnName, ok := serializationToFunctionName(fields.Serialization.Type)
			if !ok {
				continue
			}
			if !yield(&param{
				Name:          fields.Name,
				Serialization: fields.Serialization.Name,
				FunctionName:  fnName,
			}) {
				return
			}
		}
	}
}

type param struct {
	Name          string
	Serialization string
	FunctionName  string
}

var bufferPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func templateToString(t *template.Template, data any) (string, error) {
	b := bufferPool.Get().(*bytes.Buffer)
	b.Reset()
	defer bufferPool.Put(b)
	
	err := t.Execute(b, data)
	if err != nil {
		return "", fmt.Errorf("template parsing failed, template: %s, data: %+v, err: %w", t.DefinedTemplates(), data, err)
	}
	return b.String(), nil
}

func generateParams(body *repr.Data) iter.Seq[string] {
	t := getParamsTemplate()

	return func(yield func(x string) bool) {
		for p := range toSimplifiedParams(body) {
			s, err := templateToString(t, p)
			if err != nil {
				panic(err)
			}
			if !yield(s) {
				return
			}
		}

	}
}

const endpointTemplate = `app.{{ .Method | httpMethodToFiber }}(
		"{{ .Path | pathToString }}",
		{{ range formatMiddleware .Middleware .Imports }}{{.}},
		{{ end }}func(c *fiber.Ctx) error {
			body := &{{ .ImportIdent }}.{{ .Body.Name }}{}

			{{ if not .IsGet }}
			if err := c.BodyParser(body); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(struct{Err string}{Err: "Bad request"})
			}
			{{ end }}
	
			{{ range .Body | generateParams }}
			{{.}}{{end}}
	
			err := v.Struct(body)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(struct{Err string}{Err: "Validation failed"})
			}
	
			res, err := {{ .RecieverIdent }}.{{ .Handler.Name }}(
				c.UserContext(),
				body,
			)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(struct{ Err string }{Err: "InternalServerError"})
			}
	
			return c.JSON(res)
		},
	)
	`

const registerTemplate = `// Code generated by apispec. DO NOT EDIT.

package apispec

import (
	"github.com/gofiber/fiber/v2"
	"{{.ValidateImport}}"

	{{ range .Imports}}
	{{.Ident}} "{{.Import}}"{{end}}
)

func RegisterHandlers(
	app *fiber.App,
	v *validate.Validate,
	{{ range .Recievers}}{{ .Ident }} {{ .Pointer | isPointer }}{{ .Import | importIdent }}.{{ .Name }},
	{{end}}
) {
	{{ range .Endpoints }}{{.}}{{end}}
}
`
