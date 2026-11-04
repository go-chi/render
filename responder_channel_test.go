package render_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/render"
)

func TestDefaultResponderChannelTimeout(t *testing.T) {
	for _, accept := range []string{"application/json", "application/xml"} {
		t.Run(accept, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
			req.Header.Set("Accept", accept)
			render.Status(req, http.StatusAccepted)
			rec := httptest.NewRecorder()

			render.DefaultResponder(rec, req, make(chan string))

			resp := rec.Result()
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusGatewayTimeout {
				t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusGatewayTimeout)
			}
			if got := resp.Header.Get("Content-Type"); got != "text/plain; charset=utf-8" {
				t.Errorf("Content-Type = %q, want text/plain; charset=utf-8", got)
			}
			if got := rec.Body.String(); got != "Server Timeout\n" {
				t.Errorf("body = %q, want %q", got, "Server Timeout\n")
			}
		})
	}
}

func TestDefaultResponderCompletedChannel(t *testing.T) {
	for _, tt := range []struct {
		name   string
		values []string
		body   string
	}{
		{name: "empty", body: "null\n"},
		{name: "buffered", values: []string{"ready", "done"}, body: "[\"ready\",\"done\"]\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			stream := make(chan string, len(tt.values))
			for _, value := range tt.values {
				stream <- value
			}
			close(stream)
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept", "application/json")
			rec := httptest.NewRecorder()

			render.DefaultResponder(rec, req, stream)

			resp := rec.Result()
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
			}
			if got := rec.Body.String(); got != tt.body {
				t.Errorf("body = %q, want %q", got, tt.body)
			}
		})
	}
}
