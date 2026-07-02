package http

import (
	"reflect"
	"strings"
	"testing"
)

type testUserData struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type testWithTags struct {
	Token  string `json:"token"`
	Cookie string `json:"-" as:"cookie,-"`
}

func TestJR_ReflectDataField(t *testing.T) {
	fn := func() *JR[testUserData] { return nil }
	fnType := reflect.TypeOf(fn)

	retType := fnType.Out(0)
	if retType.Kind() != reflect.Pointer {
		t.Fatalf("expected pointer, got %v", retType.Kind())
	}

	elem := retType.Elem()
	if elem.Kind() != reflect.Struct {
		t.Fatalf("expected struct, got %v", elem.Kind())
	}

	if !strings.HasPrefix(elem.Name(), "JR[") {
		t.Fatalf("expected name to start with JR[, got %s", elem.Name())
	}

	dataField, ok := elem.FieldByName("Data")
	if !ok {
		t.Fatal("expected Data field to exist")
	}

	dataType := dataField.Type
	if dataType.Kind() != reflect.Struct {
		t.Fatalf("expected Data field to be struct, got %v", dataType.Kind())
	}

	if dataType.Name() != "testUserData" {
		t.Fatalf("expected Data type name testUserData, got %s", dataType.Name())
	}

	if dataType.NumField() != 2 {
		t.Fatalf("expected 2 fields in Data, got %d", dataType.NumField())
	}

	nameField := dataType.Field(0)
	if nameField.Name != "Name" || nameField.Type.Kind() != reflect.String {
		t.Fatalf("expected first field Name string, got %s %v", nameField.Name, nameField.Type.Kind())
	}
}

func TestHR_ReflectDataField(t *testing.T) {
	fn := func() *HR[testUserData] { return nil }
	fnType := reflect.TypeOf(fn)

	retType := fnType.Out(0)
	elem := retType.Elem()

	if !strings.HasPrefix(elem.Name(), "HR[") {
		t.Fatalf("expected name to start with HR[, got %s", elem.Name())
	}

	// Verify Template field exists and is any (interface{})
	templateField, ok := elem.FieldByName("Template")
	if !ok {
		t.Fatal("expected Template field to exist")
	}
	if templateField.Type.Kind() != reflect.Interface {
		t.Fatalf("expected Template field to be interface, got %v", templateField.Type.Kind())
	}

	// Verify Status field
	statusField, ok := elem.FieldByName("Status")
	if !ok {
		t.Fatal("expected Status field to exist")
	}
	if statusField.Type.Kind() != reflect.Uint16 {
		t.Fatalf("expected Status field to be uint16, got %v", statusField.Type.Kind())
	}

	// Verify Data field resolves to inner type
	dataField, ok := elem.FieldByName("Data")
	if !ok {
		t.Fatal("expected Data field to exist")
	}
	if dataField.Type.Name() != "testUserData" {
		t.Fatalf("expected Data type name testUserData, got %s", dataField.Type.Name())
	}
}

func TestJR_PkgPath(t *testing.T) {
	fn := func() *JR[testUserData] { return nil }
	fnType := reflect.TypeOf(fn)

	retType := fnType.Out(0).Elem()
	pkgPath := retType.PkgPath()

	if !strings.HasSuffix(pkgPath, "pkg/http") {
		t.Fatalf("expected PkgPath to end with pkg/http, got %s", pkgPath)
	}
}

func TestJR_DataWithTags(t *testing.T) {
	fn := func() *JR[testWithTags] { return nil }
	fnType := reflect.TypeOf(fn)

	retType := fnType.Out(0).Elem()
	dataField, _ := retType.FieldByName("Data")
	dataType := dataField.Type

	cookieField, ok := dataType.FieldByName("Cookie")
	if !ok {
		t.Fatal("expected Cookie field on data type")
	}

	asTag := cookieField.Tag.Get("as")
	if asTag != "cookie,-" {
		t.Fatalf("expected as tag 'cookie,-', got '%s'", asTag)
	}
}

func TestHR_DataWithTags(t *testing.T) {
	fn := func() *HR[testWithTags] { return nil }
	fnType := reflect.TypeOf(fn)

	retType := fnType.Out(0).Elem()
	dataField, _ := retType.FieldByName("Data")
	dataType := dataField.Type

	cookieField, ok := dataType.FieldByName("Cookie")
	if !ok {
		t.Fatal("expected Cookie field on data type")
	}

	asTag := cookieField.Tag.Get("as")
	if asTag != "cookie,-" {
		t.Fatalf("expected as tag 'cookie,-', got '%s'", asTag)
	}
}
