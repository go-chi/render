package render

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetContentType(t *testing.T) {
	tests := []struct {
		input string
		want  ContentType
	}{
		{"text/plain", ContentTypePlainText},
		{"text/html", ContentTypeHTML},
		{"application/xhtml+xml", ContentTypeHTML},
		{"application/json", ContentTypeJSON},
		{"text/javascript", ContentTypeJSON},
		{"text/xml", ContentTypeXML},
		{"application/xml", ContentTypeXML},
		{"application/x-www-form-urlencoded", ContentTypeForm},
		{"text/event-stream", ContentTypeEventStream},
		{"application/json; charset=utf-8", ContentTypeJSON},
		{"unknown/type", ContentTypeUnknown},
	}

	for _, tt := range tests {
		got := GetContentType(tt.input)
		if got != tt.want {
			t.Errorf("GetContentType(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestGetAcceptedContentType(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "application/json, text/plain")

	if got := GetAcceptedContentType(req); got != ContentTypeJSON {
		t.Errorf("GetAcceptedContentType() = %v, want %v", got, ContentTypeJSON)
	}

	reqEmpty := httptest.NewRequest("GET", "/", nil)
	if got := GetAcceptedContentType(reqEmpty); got != ContentTypePlainText {
		t.Errorf("GetAcceptedContentType() for empty Accept = %v, want %v", got, ContentTypePlainText)
	}
}

func TestSetContentType(t *testing.T) {
	handler := SetContentType(ContentTypeJSON)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := GetRequestContentType(r); got != ContentTypeJSON {
			t.Errorf("GetRequestContentType inside handler = %v, want %v", got, ContentTypeJSON)
		}
		if got := GetAcceptedContentType(r); got != ContentTypeJSON {
			t.Errorf("GetAcceptedContentType inside handler = %v, want %v", got, ContentTypeJSON)
		}
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
}
