package render

import (
	"bytes"
	"net/http"
	"strings"
	"testing"
)

func TestDecodeJSONRejectsTrailingData(t *testing.T) {
	var v map[string]interface{}
	err := DecodeJSON(strings.NewReader("{}\n some garbage data"), &v)
	if err == nil {
		t.Fatal("expected error for trailing garbage after JSON object")
	}
}

func TestDecodeJSONAllowsTrailingWhitespace(t *testing.T) {
	var v map[string]string
	if err := DecodeJSON(strings.NewReader("{\"a\":\"b\"}\n\t "), &v); err != nil {
		t.Fatal(err)
	}
	if v["a"] != "b" {
		t.Fatalf("got %#v", v)
	}
}

func TestDecodeJSONValidObject(t *testing.T) {
	var v struct {
		N int `json:"n"`
	}
	if err := DecodeJSON(bytes.NewReader([]byte(`{"n":7}`)), &v); err != nil {
		t.Fatal(err)
	}
	if v.N != 7 {
		t.Fatalf("got %d", v.N)
	}
}

func TestDefaultDecoderJSONTrailingData(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "/", strings.NewReader("{}\n extra"))
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/json")
	var v map[string]interface{}
	if err := DefaultDecoder(r, &v); err == nil {
		t.Fatal("expected trailing-data error")
	}
}
