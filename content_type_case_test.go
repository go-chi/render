package render

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetContentTypeCase(t *testing.T) {
	for _, tc := range []struct {
		name string
		want ContentType
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
	} {
		for _, value := range []string{tc.name, strings.ToUpper(tc.name), strings.ToUpper(tc.name[:1]) + tc.name[1:]} {
			for _, params := range []string{"", "; charset=UTF-8"} {
				input := " \t" + value + params + " \t"
				t.Run(input, func(t *testing.T) {
					if got := GetContentType(input); got != tc.want {
						t.Fatalf("GetContentType(%q) = %v, want %v", input, got, tc.want)
					}
				})
			}
		}
	}
}

func TestGetContentTypeUnsupported(t *testing.T) {
	for _, input := range []string{
		"", "APPLICATION/OCTET-STREAM", "Application/Jsonp", "application/problem+json", "application/ json",
		"Appl\u0130cation/JSON", "text/pla\u0130n", "text/javascr\u0130pt", "text/html, application/json",
	} {
		t.Run(input, func(t *testing.T) {
			if got := GetContentType(input); got != ContentTypeUnknown {
				t.Fatalf("GetContentType(%q) = %v, want unknown", input, got)
			}
		})
	}
}

func TestDefaultDecoderMediaTypeCase(t *testing.T) {
	for _, tc := range []struct {
		name, contentType, body string
	}{
		{"json", "Application/JSON; charset=UTF-8", `{"name":"case"}`},
		{"xml", "APPLICATION/XML", `<payload><name>case</name></payload>`},
		{"text xml", "Text/XML; charset=utf-8", `<payload><name>case</name></payload>`},
		{"form", "Application/X-WWW-Form-Urlencoded", "name=case"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", tc.contentType)
			var payload struct {
				Name string `json:"name" xml:"name" form:"name"`
			}
			if err := DefaultDecoder(req, &payload); err != nil {
				t.Fatal(err)
			}
			if payload.Name != "case" {
				t.Fatalf("decoded name = %q, want case", payload.Name)
			}
		})
	}
}

func TestDefaultResponderMediaTypeCase(t *testing.T) {
	for _, accept := range []string{"APPLICATION/XML", "Text/XML; charset=UTF-8", "Application/XML, application/json"} {
		t.Run(accept, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept", accept)
			payload := struct {
				XMLName xml.Name `xml:"payload"`
				Name    string   `xml:"name"`
			}{Name: "case"}
			w := httptest.NewRecorder()
			DefaultResponder(w, req, payload)
			if got := w.Header().Get("Content-Type"); got != "application/xml; charset=utf-8" {
				t.Fatalf("response Content-Type = %q, want XML", got)
			}
			var result struct {
				Name string `xml:"name"`
			}
			if err := xml.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.Name != "case" {
				t.Fatalf("response name = %q, want case", result.Name)
			}
		})
	}
}

func TestMediaTypeContextOverride(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"case"}`))
	req.Header.Set("Content-Type", "APPLICATION/XML")
	req.Header.Set("Accept", "APPLICATION/XML")
	req = req.WithContext(context.WithValue(req.Context(), ContentTypeCtxKey, ContentTypeJSON))
	if got := GetRequestContentType(req); got != ContentTypeJSON {
		t.Fatalf("request ContentType = %v, want JSON context", got)
	}
	if got := GetAcceptedContentType(req); got != ContentTypeJSON {
		t.Fatalf("accepted ContentType = %v, want JSON context", got)
	}
	var payload struct {
		Name string `json:"name"`
	}
	if err := DefaultDecoder(req, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Name != "case" {
		t.Fatalf("decoded name = %q, want case", payload.Name)
	}
}
