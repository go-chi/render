package render_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/render"
)

type childRenderer struct{}

func (childRenderer) Render(http.ResponseWriter, *http.Request) error { return nil }

type parentRenderer struct {
	child childRenderer
}

func (parentRenderer) Render(http.ResponseWriter, *http.Request) error { return nil }

type childBinder struct{}

func (childBinder) Bind(*http.Request) error { return nil }

type parentBinder struct {
	child childBinder
}

func (parentBinder) Bind(*http.Request) error { return nil }

func TestRenderSkipsUnexportedFields(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	if err := render.Render(w, r, parentRenderer{}); err != nil {
		t.Fatal(err)
	}
}

func TestBindSkipsUnexportedFields(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
	r.Header.Set("Content-Type", "application/json")
	if err := render.Bind(r, &parentBinder{}); err != nil {
		t.Fatal(err)
	}
}
