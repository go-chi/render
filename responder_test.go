package render

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestData_DefaultContentType(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	payload := []byte("binary data")

	Data(w, r, payload)

	if got := w.Header().Get("Content-Type"); got != "application/octet-stream" {
		t.Fatalf("expected Content-Type application/octet-stream, got %q", got)
	}

	if !bytes.Equal(w.Body.Bytes(), payload) {
		t.Fatalf("expected body %q, got %q", payload, w.Body.Bytes())
	}

	if w.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestData_CustomContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		payload     []byte
	}{
		{
			name:        "image/png",
			contentType: "image/png",
			payload:     []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A},
		},
		{
			name:        "application/pdf",
			contentType: "application/pdf",
			payload:     []byte("%PDF-1.4 sample pdf content"),
		},
		{
			name:        "custom text/csv with charset",
			contentType: "text/csv; charset=utf-8",
			payload:     []byte("id,name\n1,test\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			w.Header().Set("Content-Type", tt.contentType)
			Data(w, r, tt.payload)

			if got := w.Header().Get("Content-Type"); got != tt.contentType {
				t.Fatalf("expected Content-Type %q, got %q", tt.contentType, got)
			}

			if !bytes.Equal(w.Body.Bytes(), tt.payload) {
				t.Fatalf("expected body %q, got %q", tt.payload, w.Body.Bytes())
			}
		})
	}
}

func TestData_WithStatus(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	Status(r, http.StatusAccepted)

	w.Header().Set("Content-Type", "application/zip")
	payload := []byte("zipdata")

	Data(w, r, payload)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected status code %d, got %d", http.StatusAccepted, w.Code)
	}

	if got := w.Header().Get("Content-Type"); got != "application/zip" {
		t.Fatalf("expected Content-Type application/zip, got %q", got)
	}

	if !bytes.Equal(w.Body.Bytes(), payload) {
		t.Fatalf("expected body %q, got %q", payload, w.Body.Bytes())
	}
}

func TestData_EmptyPayload(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	Data(w, r, []byte{})

	if got := w.Header().Get("Content-Type"); got != "application/octet-stream" {
		t.Fatalf("expected Content-Type application/octet-stream, got %q", got)
	}

	if len(w.Body.Bytes()) != 0 {
		t.Fatalf("expected empty body, got %q", w.Body.Bytes())
	}
}
