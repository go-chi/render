package render

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type bindBody struct {
	Name     string   `json:"name"`
	Children Children `json:"children"`
}

func (b *bindBody) Bind(r *http.Request) error {
	if b.Name == "" {
		return errors.New("empty body name")
	}
	return nil
}

type Children []Child

func (c Children) Bind(r *http.Request) error {
	if c == nil {
		return errors.New("no children provided")
	}
	return nil
}

type Child struct {
	Name string `json:"name"`
}

func (c *Child) Bind(r *http.Request) error {
	if c.Name == "" {
		return errors.New("empty child name")
	}
	return nil
}

type bindPlainSlice struct {
	Kids []Child `json:"kids"`
}

func (b *bindPlainSlice) Bind(r *http.Request) error { return nil }

type translatedField map[string]string

func (t translatedField) Bind(r *http.Request) error {
	if len(t) == 0 {
		return errors.New("empty translation")
	}
	return nil
}

type bindMapField struct {
	Title translatedField `json:"title"`
}

func (b *bindMapField) Bind(r *http.Request) error { return nil }

type mapValueChild struct {
	Items map[string]Child `json:"items"`
}

func (m *mapValueChild) Bind(r *http.Request) error { return nil }

func TestBindWalksSliceElements(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
		`{"name":"ok","children":[{"name":"a"},{"name":""}]}`,
	))
	r.Header.Set("Content-Type", "application/json")
	var body bindBody
	err := Bind(r, &body)
	if err == nil || err.Error() != "empty child name" {
		t.Fatalf("got %v, want empty child name", err)
	}
}

func TestBindNamedSliceType(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
		`{"name":"ok","children":[{"name":"a"}]}`,
	))
	r.Header.Set("Content-Type", "application/json")
	var body bindBody
	if err := Bind(r, &body); err != nil {
		t.Fatal(err)
	}
}

func TestBindNilNamedSlice(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"ok"}`))
	r.Header.Set("Content-Type", "application/json")
	var body bindBody
	err := Bind(r, &body)
	if err == nil || err.Error() != "no children provided" {
		t.Fatalf("got %v, want no children provided", err)
	}
}

func TestBindUnnamedSliceElements(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
		`{"kids":[{"name":""}]}`,
	))
	r.Header.Set("Content-Type", "application/json")
	var body bindPlainSlice
	err := Bind(r, &body)
	if err == nil || err.Error() != "empty child name" {
		t.Fatalf("got %v, want empty child name", err)
	}
}

func TestBindMapTypeField(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"title":{}}`))
	r.Header.Set("Content-Type", "application/json")
	var body bindMapField
	err := Bind(r, &body)
	if err == nil || err.Error() != "empty translation" {
		t.Fatalf("got %v, want empty translation", err)
	}
}

func TestBindMapValues(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
		`{"items":{"a":{"name":""}}}`,
	))
	r.Header.Set("Content-Type", "application/json")
	var body mapValueChild
	err := Bind(r, &body)
	if err == nil || err.Error() != "empty child name" {
		t.Fatalf("got %v, want empty child name", err)
	}
}
