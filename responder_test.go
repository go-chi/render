package render

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSON(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	JSON(w, r, obj)

	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if body := w.Body.String(); body != str {
		t.Errorf("handler returned wrong body: got %v want %v", body, str)
	}
}

func BenchmarkJSON(b *testing.B) {
	b.ReportAllocs()

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	for n := 0; n < b.N; n++ {
		w.Body.Reset()
		JSON(w, r, obj)
	}
}

type (
	Obj struct {
		ID      int64
		Name    string
		Words   []string
		Numbers []float64
	}
)

var (
	obj = Obj{
		ID:      123,
		Name:    "Some data",
		Words:   []string{"one", "two", "three", "four", "five"},
		Numbers: []float64{1.0, 2.0, 3.0, 4.0, 5.0},
	}
	str = `{"ID":123,"Name":"Some data","Words":["one","two","three","four","five"],"Numbers":[1,2,3,4,5]}
`
)
