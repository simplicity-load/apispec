package generate_test

import (
	"bytes"
	"context"
	"testing"

	generate "github.com/simplicity-load/apispec/pkg/gen"
	"github.com/simplicity-load/apispec/pkg/http"
	"github.com/simplicity-load/apispec/pkg/parse/server"
)

type testHandler struct{}

type TestReq struct {
	Y string `json:"y"`
}

type TestRes struct {
	Y string `json:"y"`
}

func (h testHandler) Get(ctx context.Context, param *TestReq) (*http.JR[TestRes], error) {
	return nil, nil
}
func (h testHandler) Post(ctx context.Context, param *TestReq) (*http.JR[TestRes], error) {
	return nil, nil
}

func TestGenerate(t *testing.T) {
	h := testHandler{}
	app := http.NewAPI()
	app.Get(h.Get, "desc")
	app.Post(h.Post, "desc")
	sus := app.Static("sus")
	sus.Get(h.Get, "desc")
	wus := sus.Param("wus")
	wus.Get(h.Get, "desc")
	paths, err := server.ParsePaths(app)
	if err != nil {
		t.Fatalf("ParsePaths failed: %v", err)
	}

	var buf bytes.Buffer
	err = generate.Generate(paths, &buf, "")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("Generate produced empty output")
	}
}
