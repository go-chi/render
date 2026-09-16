package render

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJSONNilRequest(t *testing.T) {
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusCreated)
	JSON(w, nil, map[string]string{"field": "test"})
	if w.Code != http.StatusCreated {
		t.Fatalf("status=%d want %d", w.Code, http.StatusCreated)
	}
	got := strings.TrimSpace(w.Body.String())
	if got != `{"field":"test"}` {
		t.Fatalf("body=%q", got)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type=%q", ct)
	}
}

func TestXMLNilRequest(t *testing.T) {
	type payload struct {
		XMLName xml.Name `xml:"item"`
		Field   string   `xml:"field"`
	}
	w := httptest.NewRecorder()
	XML(w, nil, payload{Field: "test"})
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d want %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), "<field>test</field>") {
		t.Fatalf("body=%q", w.Body.String())
	}
}

func TestPlainTextDataHTMLNilRequest(t *testing.T) {
	w := httptest.NewRecorder()
	PlainText(w, nil, "hello")
	if w.Body.String() != "hello" {
		t.Fatalf("plaintext body=%q", w.Body.String())
	}

	w = httptest.NewRecorder()
	Data(w, nil, []byte("raw"))
	if w.Body.String() != "raw" {
		t.Fatalf("data body=%q", w.Body.String())
	}

	w = httptest.NewRecorder()
	HTML(w, nil, "<p>ok</p>")
	if w.Body.String() != "<p>ok</p>" {
		t.Fatalf("html body=%q", w.Body.String())
	}
}

func TestJSONStillHonorsStatusHint(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	Status(r, http.StatusAccepted)
	JSON(w, r, map[string]string{"ok": "yes"})
	if w.Code != http.StatusAccepted {
		t.Fatalf("status=%d want %d", w.Code, http.StatusAccepted)
	}
}
