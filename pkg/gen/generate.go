package generate

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"strings"
	"sync"
	"text/template"

	"github.com/simplicity-load/apispec/pkg/http"
	repr "github.com/simplicity-load/apispec/pkg/repr/http"
)

//go:embed gofiber/fiber.tmpl
var templ string

func getRegisterTemplate(imports importSet[sorted], recievers recieverSet[sorted]) *template.Template {
	return sync.OnceValue(func() *template.Template {
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
			"formatMiddleware": func(middleware repr.Middlewares) []string {
				return formatMiddleware(middleware, imports, recievers)
			},
			"httpMethodToFnIdent": httpMethodToFiber,
			"pathToString":        repr.PathToURL,
			"toRequestParams":     toRequestParams,
			"toRespParams":        toRespParams,
		}).Parse(templ)
		if err != nil {
			panic(err)
		}
		return t
	})()
}

func Generate(
	representation repr.Representation,
	output io.Writer,
	validateUrl string,
) error {
	imports := newImportSet()
	recievers := newRecieverSet()
	endpoints := generateEndpoints(representation.Routes, imports, recievers)

	impSortSet, impSort := imports.sort()
	recvSortSet, recvSort := recievers.sort()

	t := getRegisterTemplate(impSortSet, recvSortSet)

	enrinchedEndpoints := make([]endpointTemplateData, len(endpoints))
	for i, e := range endpoints {
		recvIdent := recvSortSet.get(e.Handler.Reciever, e.Handler.Import)
		impIdent := impSortSet.get(e.Body.Import)
		enrinchedEndpoints[i] = endpointTemplateData{
			AppIdent:      "app",
			Endpoint:      e.Endpoint,
			IsGet:         e.IsGet,
			Middleware:    e.Middleware,
			RecieverIdent: recvIdent,
			ImportIdent:   impIdent,
		}
	}

	data := struct {
		Recievers      []reciever
		Imports        []importer
		Endpoints      []endpointTemplateData
		ValidateImport string
		SetupImports   []string
	}{
		Recievers:      recvSort,
		Imports:        impSort,
		Endpoints:      enrinchedEndpoints,
		ValidateImport: validateUrl,
		SetupImports: []string{
			"github.com/gofiber/fiber/v2",
			validateUrl,
		},
	}
	gen, err := templateToString(t.Lookup("setup"), data)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(output, gen)
	if err != nil {
		return err
	}
	return nil
}

func generateEndpoints(
	path *repr.Path,
	imports importSet[unsorted],
	recievers recieverSet[unsorted],
) []endpointTemplateData {
	return generateEndpointsIter(path, imports, recievers)
}

func generateEndpointsIter(
	path *repr.Path,
	imports importSet[unsorted],
	recievers recieverSet[unsorted],
) []endpointTemplateData {
	for imp := range path.Endpoints.Imports() {
		imports.add(imp)
	}
	for imp := range path.Middleware.Imports() {
		imports.add(imp)
	}
	for reciever, imp := range path.Endpoints.Recievers() {
		recievers.add(reciever, imp)
	}

	endpointAcc := make([]endpointTemplateData, len(path.Endpoints))
	for i, e := range path.Endpoints {
		endpointAcc[i] = endpointTemplateData{
			Endpoint:   e,
			IsGet:      e.Method == http.GET,
			Middleware: path.Middleware,
		}
	}

	for _, subPath := range path.SubPath {
		endpoints := generateEndpointsIter(subPath, imports, recievers)
		endpointAcc = append(endpointAcc, endpoints...)
	}
	return endpointAcc
}

type endpointTemplateData struct {
	*repr.Endpoint
	IsGet         bool
	ImportIdent   string
	RecieverIdent string
	Middleware    repr.Middlewares
	AppIdent      string
}

func httpMethodToFiber(method http.Method) string {
	x := string(method)
	return strings.ToUpper(x[:1]) + strings.ToLower(x[1:])
}

func formatMiddleware(
	middleware repr.Middlewares,
	imports importSet[sorted],
	recievers recieverSet[sorted],
) []string {
	formatted := make([]string, len(middleware))
	for i, m := range middleware {
		if m.Reciever == nil {
			formatted[i] = fmt.Sprintf("%s.%s",
				imports.get(m.Import),
				m.Name)
		} else {
			formatted[i] = fmt.Sprintf("%s.%s",
				recievers.get(m.Reciever, m.Import),
				m.Name)
		}
	}
	return formatted
}

func toRequestParams(data *repr.Data) []*param {
	return toParams(serializationToRequestFiber)(data)
}
func serializationToRequestFiber(t repr.SerializationType) (string, bool) {
	fnName, ok := map[repr.SerializationType]string{
		repr.SerializationPATH:   "c.Params",
		repr.SerializationQUERY:  "c.Query",
		repr.SerializationHEADER: "c.Get",
	}[t]
	return fnName, ok
}

func toRespParams(data *repr.Data) []*param {
	return toParams(serializationToRespFiber)(data)
}
func serializationToRespFiber(t repr.SerializationType) (string, bool) {
	fnName, ok := map[repr.SerializationType]string{
		repr.SerializationHEADER: "c.Set",
		repr.SerializationCOOKIE: "c.Set",
	}[t]
	return fnName, ok
}

func toParams(
	fn func(repr.SerializationType) (string, bool),
) func(data *repr.Data) []*param {
	return func(data *repr.Data) []*param {
		params := make([]*param, 0, len(data.Fields))
		for _, fields := range data.Fields {
			fnName, ok := fn(fields.Serialization.Type)
			if !ok {
				continue
			}

			serializationName := fields.Serialization.Name
			if fields.Serialization.Type ==
				repr.SerializationCOOKIE {
				serializationName = "set-cookie"
			}

			params = append(params, &param{
				Name:          fields.Name,
				Serialization: serializationName,
				FunctionName:  fnName,
			})
		}
		return params
	}
}

type param struct {
	Name          string
	Serialization string
	FunctionName  string
}

var bufferPool = sync.Pool{
	New: func() any { return bytes.NewBuffer(nil) },
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
