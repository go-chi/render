package render

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type (
	profile struct {
		Name string
		Age  int
	}
	publicProfile profile
	profileJSON   struct {
		Name string `json:"name"`
	}
	prepareOnly struct {
		Name string `json:"name"`
	}
)

func (p publicProfile) Render(w http.ResponseWriter, r *http.Request) error {
	Status(r, http.StatusNoContent)
	JSON(w, r, profileJSON{Name: p.Name})
	return nil
}

func (p prepareOnly) Render(http.ResponseWriter, *http.Request) error {
	return nil
}

func TestRenderSkipsDefaultWhenRendererWrote(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	if err := Render(w, r, publicProfile{Name: "John", Age: 42}); err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(w.Body)
	got := strings.TrimSpace(string(body))
	if w.Code != http.StatusNoContent {
		t.Fatalf("status=%d want %d", w.Code, http.StatusNoContent)
	}
	if got != `{"name":"John"}` {
		t.Fatalf("body=%q, extra default JSON dump?", got)
	}
	if strings.Contains(got, `"Age"`) {
		t.Fatalf("leaked aliased fields: %s", got)
	}
}

func TestRenderStillSerializesPrepareOnlyRenderer(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	if err := Render(w, r, prepareOnly{Name: "Ada"}); err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(w.Body)
	got := strings.TrimSpace(string(body))
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d want %d", w.Code, http.StatusOK)
	}
	if got != `{"name":"Ada"}` {
		t.Fatalf("body=%q", got)
	}
}

func TestRenderListSkipsDefaultWhenRendererWrote(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	list := []Renderer{
		publicProfile{Name: "John", Age: 42},
	}
	if err := RenderList(w, r, list); err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(w.Body)
	got := strings.TrimSpace(string(body))
	if got != `{"name":"John"}` {
		t.Fatalf("body=%q", got)
	}
}
