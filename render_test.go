package render

import (
	"bufio"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type basicRenderInterface struct {
	RenderStruct Renderer
}

func (h *basicRenderInterface) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type baseRenderInterface struct {
	PtrVal         *string
	NoRenderStruct interface{}
	RenderList     []Renderer
	RenderStruct   Renderer
}

func (h *baseRenderInterface) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type otherRenderInterface struct {
	Value int
}

func (o *otherRenderInterface) Render(w http.ResponseWriter, r *http.Request) error {
	fmt.Fprintf(w, "%d", o.Value)
	return nil
}

func TestRendererInterface(t *testing.T) {
	rv := reflect.ValueOf(newTestStruct())
	if !rv.Type().Implements(rendererType) {
		t.Fatal("test element should implement Renderer interface")
	}
}

func TestBasicRenderer(t *testing.T) {
	h := &basicRenderInterface{
		RenderStruct: &otherRenderInterface{Value: 2},
	}

	// declare the request
	r := req(t, postReq("test"))

	// create the response recorder
	w := httptest.NewRecorder()

	err := renderer(w, r, h)
	if err != nil {
		t.Errorf("error encountered: %s", err)
	}

	response := fmt.Sprintf("%s", w.Body)
	if response != "2" {
		t.Errorf("unexpected response %s; expected 2", response)
	}
}

func TestRenderer(t *testing.T) {
	h := newTestStruct()

	// declare the request
	r := req(t, postReq("test"))

	// create the response recorder
	w := httptest.NewRecorder()

	err := renderer(w, r, h)
	if err != nil {
		t.Errorf("error encountered: %s", err)
	}

	response := fmt.Sprintf("%s", w.Body)
	if response != "123" {
		t.Errorf("unexpected response %s; expected 123", response)
	}
}

func newTestStruct() *baseRenderInterface {
	return &baseRenderInterface{
		PtrVal:         nil,
		NoRenderStruct: otherRenderInterface{},
		RenderList: []Renderer{
			&otherRenderInterface{Value: 1},
			&otherRenderInterface{Value: 2},
		},
		RenderStruct: &otherRenderInterface{Value: 3},
	}
}

func req(t testing.TB, v string) *http.Request {
	req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(v)))
	if err != nil {
		t.Fatal(err)
	}
	return req
}

func postReq(cont string) string {
	post :=
		`POST / HTTP/1.1
Content-Type: application/vnd.api+json
User-Agent: mockagent
Content-Length: %d

%s`
	return fmt.Sprintf(post, len(cont), cont)
}
