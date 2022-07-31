package render

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

type example struct {
	Field string `json:"field"`
}

func createTestInput() io.Reader {
	return bytes.NewBuffer([]byte(`{"field":"some input","unknown":null}`))
}

func TestStrictDecoder(t *testing.T) {
	ex := &example{}

	t.Run("strict", func(t *testing.T) {
		// Test strict
		buf := createTestInput()
		err := DecodeJSON(buf, ex, true)
		// Could not find a custom error type for disallowed fields
		// https://github.com/golang/go/issues/40982
		if !strings.Contains(err.Error(), "unknown field") || err == nil {
			t.Errorf("should return error and contain 'unkown field'\n")
		}
	})

	t.Run("relaxed", func(t *testing.T) {
		// Test relaxed
		buf := createTestInput()
		err := DecodeJSON(buf, ex)
		// Make sure strict is false by default
		if err != nil {
			t.Errorf("should not return error: %v\n", err)
		}
	})

}
