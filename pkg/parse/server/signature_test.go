package server

import (
	"context"
	"reflect"
	"testing"

	"github.com/simplicity-load/apispec/pkg/http"
	repr "github.com/simplicity-load/apispec/pkg/repr/http"
)

type testReqData struct {
	Name string `json:"name" validate:"required"`
}

type testResData struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type testResWithCookie struct {
	Token  string `json:"token"`
	Cookie string `json:"-" as:"cookie,-"`
}

func TestParseFnSignature_JR(t *testing.T) {
	handler := func(ctx context.Context, req *testReqData) (*http.JR[testResData], error) {
		return nil, nil
	}

	fn := reflect.TypeOf(handler)
	reqType, resType, kind, err := parseFnSignature(fn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if kind != repr.ResponseJSON {
		t.Fatalf("expected ResponseJSON, got %s", kind)
	}

	if reqType.Name() != "testReqData" {
		t.Fatalf("expected reqType testReqData, got %s", reqType.Name())
	}

	if resType.Name() != "testResData" {
		t.Fatalf("expected resType testResData, got %s", resType.Name())
	}

	if resType.NumField() != 2 {
		t.Fatalf("expected 2 fields on resType, got %d", resType.NumField())
	}
}

func TestParseFnSignature_HR(t *testing.T) {
	handler := func(ctx context.Context, req *testReqData) (*http.HR[testResData], error) {
		return nil, nil
	}

	fn := reflect.TypeOf(handler)
	reqType, resType, kind, err := parseFnSignature(fn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if kind != repr.ResponseHTML {
		t.Fatalf("expected ResponseHTML, got %s", kind)
	}

	if reqType.Name() != "testReqData" {
		t.Fatalf("expected reqType testReqData, got %s", reqType.Name())
	}

	if resType.Name() != "testResData" {
		t.Fatalf("expected resType testResData, got %s", resType.Name())
	}
}

func TestParseFnSignature_JRWithCookieTags(t *testing.T) {
	handler := func(ctx context.Context, req *testReqData) (*http.JR[testResWithCookie], error) {
		return nil, nil
	}

	fn := reflect.TypeOf(handler)
	_, resType, kind, err := parseFnSignature(fn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if kind != repr.ResponseJSON {
		t.Fatalf("expected ResponseJSON, got %s", kind)
	}

	cookieField, ok := resType.FieldByName("Cookie")
	if !ok {
		t.Fatal("expected Cookie field on resType")
	}

	asTag := cookieField.Tag.Get("as")
	if asTag != "cookie,-" {
		t.Fatalf("expected as tag 'cookie,-', got '%s'", asTag)
	}
}

func TestParseFnSignature_PlainStructReturnsError(t *testing.T) {
	handler := func(ctx context.Context, req *testReqData) (*testResData, error) {
		return nil, nil
	}

	fn := reflect.TypeOf(handler)
	_, _, _, err := parseFnSignature(fn)
	if err == nil {
		t.Fatal("expected error for plain struct return, got nil")
	}
}

func TestParseFnSignature_WrongParamCount(t *testing.T) {
	handler := func(ctx context.Context) (*http.JR[testResData], error) {
		return nil, nil
	}

	fn := reflect.TypeOf(handler)
	_, _, _, err := parseFnSignature(fn)
	if err == nil {
		t.Fatal("expected error for wrong param count, got nil")
	}
}

func TestParseFnSignature_WrongReturnCount(t *testing.T) {
	handler := func(ctx context.Context, req *testReqData) *http.JR[testResData] {
		return nil
	}

	fn := reflect.TypeOf(handler)
	_, _, _, err := parseFnSignature(fn)
	if err == nil {
		t.Fatal("expected error for wrong return count, got nil")
	}
}
