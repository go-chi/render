package render_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/render"
)

type valueChildRenderer struct{}

func (valueChildRenderer) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type parentWithPrivateRenderer struct {
	child valueChildRenderer
	tag   string
}

func (p parentWithPrivateRenderer) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type valueChildBinder struct{}

func (valueChildBinder) Bind(r *http.Request) error {
	return nil
}

type parentWithPrivateBinder struct {
	child valueChildBinder
	tag   string
}

func (p *parentWithPrivateBinder) Bind(r *http.Request) error {
	return nil
}

type exportedChildRenderer struct {
	called *bool
}

func (e exportedChildRenderer) Render(w http.ResponseWriter, r *http.Request) error {
	*e.called = true
	return nil
}

type parentWithMixedRenderer struct {
	child    valueChildRenderer
	Exported exportedChildRenderer
}

func (p parentWithMixedRenderer) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type exportedChildBinder struct {
	called *bool
}

func (e *exportedChildBinder) Bind(r *http.Request) error {
	*e.called = true
	return nil
}

type parentWithMixedBinder struct {
	child    valueChildBinder
	Exported *exportedChildBinder
}

func (p *parentWithMixedBinder) Bind(r *http.Request) error {
	return nil
}

type parentWithNilPointerRenderer struct {
	Exported *exportedChildRenderer
}

func (p parentWithNilPointerRenderer) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func TestPrivateRenderer_DoesNotPanic(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	payload := parentWithPrivateRenderer{
		child: valueChildRenderer{},
		tag:   "test",
	}

	if err := render.Render(w, r, payload); err != nil {
		t.Fatalf("unexpected error rendering payload with private field: %v", err)
	}
}

func TestPrivateBinder_DoesNotPanic(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"tag":"test"}`))
	r.Header.Set("Content-Type", "application/json")

	payload := &parentWithPrivateBinder{}
	if err := render.Bind(r, payload); err != nil {
		t.Fatalf("unexpected error binding payload with private field: %v", err)
	}
}

func TestMixedFields_ExecutesExportedAndSkipsUnexported(t *testing.T) {
	// Renderer test
	calledRenderer := false
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	mixedPayload := parentWithMixedRenderer{
		child: valueChildRenderer{},
		Exported: exportedChildRenderer{
			called: &calledRenderer,
		},
	}

	if err := render.Render(w, r, mixedPayload); err != nil {
		t.Fatalf("unexpected error rendering mixed payload: %v", err)
	}
	if !calledRenderer {
		t.Fatalf("expected Exported renderer to be executed")
	}

	// Binder test
	calledBinder := false
	rBind := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	rBind.Header.Set("Content-Type", "application/json")

	mixedBinder := &parentWithMixedBinder{
		child: valueChildBinder{},
		Exported: &exportedChildBinder{
			called: &calledBinder,
		},
	}

	if err := render.Bind(rBind, mixedBinder); err != nil {
		t.Fatalf("unexpected error binding mixed payload: %v", err)
	}
	if !calledBinder {
		t.Fatalf("expected Exported binder to be executed")
	}
}

func TestNilPointerRendererField(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	payload := parentWithNilPointerRenderer{
		Exported: nil,
	}

	if err := render.Render(w, r, payload); err != nil {
		t.Fatalf("unexpected error rendering nil pointer field: %v", err)
	}
}
