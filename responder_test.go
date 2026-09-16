package render

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type leakyJSON struct{}

func (leakyJSON) MarshalJSON() ([]byte, error) {
	return nil, errors.New("secret token xyz")
}

func TestJSONDoesNotLeakEncodeError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	JSON(rec, req, leakyJSON{})
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	body := rec.Body.String()
	if strings.Contains(body, "secret token") {
		t.Fatalf("leaked encode error: %q", body)
	}
	if !strings.Contains(body, http.StatusText(http.StatusInternalServerError)) {
		t.Fatalf("body = %q, want Internal Server Error", body)
	}
}

func TestJSONOK(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	JSON(rec, req, M{"ok": true})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Body.String(); !strings.Contains(got, `"ok":true`) {
		t.Fatalf("body = %q", got)
	}
}
