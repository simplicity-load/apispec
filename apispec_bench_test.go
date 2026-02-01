package apispec_test

import (
	"context"
	"io"
	"os"
	"strconv"
	"testing"

	"github.com/simplicity-load/apispec"
	"github.com/simplicity-load/apispec/pkg/http"
)

type X struct{}

func (x X) Get(ctx context.Context, param *struct{}) (*struct{}, error)    { return nil, nil }
func (x X) Post(ctx context.Context, param *struct{}) (*struct{}, error)   { return nil, nil }
func (x X) Put(ctx context.Context, param *struct{}) (*struct{}, error)    { return nil, nil }
func (x X) Patch(ctx context.Context, param *struct{}) (*struct{}, error)  { return nil, nil }
func (x X) Delete(ctx context.Context, param *struct{}) (*struct{}, error) { return nil, nil }

func generateApiRoutes(b *testing.B, pathCount int) *http.Path {
	b.Helper()
	x := X{}
	app := http.NewAPI()
	app.Get(x.Get, "desc")
	app.Post(x.Post, "desc")
	app.Put(x.Put, "desc")
	app.Patch(x.Patch, "desc")
	app.Delete(x.Delete, "desc")

	sus := app.Static("sus")
	sus.Get(x.Get, "desc")
	sus.Post(x.Post, "desc")
	sus.Put(x.Put, "desc")
	sus.Patch(x.Patch, "desc")
	sus.Delete(x.Delete, "desc")

	wus := sus.Param("wus")
	wus.Get(x.Get, "desc")
	wus.Post(x.Post, "desc")
	wus.Put(x.Put, "desc")
	wus.Patch(x.Patch, "desc")
	wus.Delete(x.Delete, "desc")

	for i := range pathCount {
		temp := sus.Static(strconv.FormatInt(int64(i), 10))
		temp.Get(x.Get, "desc")
		temp.Post(x.Post, "desc")
		temp.Put(x.Put, "desc")
		temp.Patch(x.Patch, "desc")
		temp.Delete(x.Delete, "desc")
	}

	return app
}

func benchmarkGenerate(b *testing.B, routeCount int) {
	os.Stdout = os.NewFile(0, os.DevNull)
	for b.Loop() {
		b.StopTimer()
		routes := generateApiRoutes(b, routeCount)
		b.StartTimer()
		err := apispec.Generate(http.HttpServer{
			Routes:     routes,
			OutputFile: io.Discard,
		})
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkGenerate1k(b *testing.B)   { benchmarkGenerate(b, 200) }
func BenchmarkGenerate10k(b *testing.B)  { benchmarkGenerate(b, 2000) }
func BenchmarkGenerate100k(b *testing.B) { benchmarkGenerate(b, 20000) }
func BenchmarkGenerate1M(b *testing.B)   { benchmarkGenerate(b, 200000) }
