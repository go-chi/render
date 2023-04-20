package render

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"reflect"
)

var ErrInvalidType error = errors.New("Invalid Type passed")

// Encoder is a generic interface to encode arbitrary data from value `v`  to a
// reader `r`
type Encoder interface {
	Encode(w http.ResponseWriter, req *http.Request, v interface{}) error
}

// Package-level variables for encoding the supported formats. They are set to
// our default implementations. By setting render.Encode{JSON,XML,Form} you can
// customize Encoding (e.g. you might want to configure the JSON-encoder)
// TODO document type constraints
var (
	EncoderJSON      Encoder = EncodeJSON{}
	EncoderXML       Encoder = EncodeXML{}
	EncoderData      Encoder = EncodeData{}
	EncoderPlainText Encoder = EncodePlainText{}
)

// M is a convenience alias for quickly building a map structure that is going
// out to a responder. Just a short-hand.
type M map[string]interface{}

// Respond is a package-level variable set to our default Responder. We do this
// because it allows you to set render.Respond to another function with the
// same function signature, while also utilizing the render.Responder() function
// itself. Effectively, allowing you to easily add your own logic to the package
// defaults. For example, maybe you want to test if v is an error and respond
// differently, or log something before you respond.
var Respond = DefaultResponder

// StatusCtxKey is a context key to record a future HTTP response status code.
var StatusCtxKey = &contextKey{"Status"}

// Status sets a HTTP response status code hint into request context at any point
// during the request life-cycle. Before the Responder sends its response header
// it will check the StatusCtxKey
func Status(r *http.Request, status int) {
	*r = *r.WithContext(context.WithValue(r.Context(), StatusCtxKey, status))
}

// Respond handles streaming JSON and XML responses, automatically setting the
// Content-Type based on request headers. It will default to a JSON response.
func DefaultResponder(w http.ResponseWriter, r *http.Request, v interface{}) {
	if v != nil {
		switch reflect.TypeOf(v).Kind() {
		case reflect.Chan:
			switch GetAcceptedContentType(r) {
			case ContentTypeEventStream:
				channelEventStream(w, r, v)
				return
			default:
				v = channelIntoSlice(w, r, v)
			}
		}
	}

	// Format response based on request Accept header.
	switch GetAcceptedContentType(r) {
	case ContentTypeJSON:
		EncoderJSON.Encode(w, r, v)
	case ContentTypeXML:
		EncoderXML.Encode(w, r, v)
	default:
		EncoderJSON.Encode(w, r, v)
	}
}

type EncodePlainText struct{}

// PlainText writes a string to the response, setting the Content-Type as
// text/plain.
func (EncodePlainText) Encode(w http.ResponseWriter, r *http.Request, vi interface{}) error {
	v, ok := vi.(string)
	if !ok {
		return ErrInvalidType
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if status, ok := r.Context().Value(StatusCtxKey).(int); ok {
		w.WriteHeader(status)
	}
	w.Write([]byte(v)) //nolint:errcheck
	return nil
}

// TODO backwards compatible
type EncodeData struct{}

// Data writes raw bytes to the response, setting the Content-Type as
// application/octet-stream.
func (EncodeData) Encode(w http.ResponseWriter, r *http.Request, vi interface{}) error {
	v, ok := vi.([]byte)
	if !ok {
		return ErrInvalidType
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	if status, ok := r.Context().Value(StatusCtxKey).(int); ok {
		w.WriteHeader(status)
	}
	w.Write(v) //nolint:errcheck
	return nil
}

// TODO backwards compatible
type EncodeHTML struct{}

// HTML writes a string to the response, setting the Content-Type as text/html.
func (EncodeHTML) Encode(w http.ResponseWriter, r *http.Request, vi interface{}) error {
	v, ok := vi.(string)
	if !ok {
		return ErrInvalidType
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if status, ok := r.Context().Value(StatusCtxKey).(int); ok {
		w.WriteHeader(status)
	}
	w.Write([]byte(v)) //nolint:errcheck
	return nil
}

type EncodeJSON struct{}

// JSON marshals 'v' to JSON, automatically escaping HTML and setting the
// Content-Type as application/json.
func (EncodeJSON) Encode(w http.ResponseWriter, r *http.Request, v interface{}) error {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	if status, ok := r.Context().Value(StatusCtxKey).(int); ok {
		w.WriteHeader(status)
	}
	w.Write(buf.Bytes()) //nolint:errcheck
	return nil
}

type EncodeXML struct{}

// XML marshals 'v' to XML, setting the Content-Type as application/xml. It
// will automatically prepend a generic XML header (see encoding/xml.Header) if
// one is not found in the first 100 bytes of 'v'.
func (EncodeXML) Encode(w http.ResponseWriter, r *http.Request, v interface{}) error {
	b, err := xml.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	if status, ok := r.Context().Value(StatusCtxKey).(int); ok {
		w.WriteHeader(status)
	}

	// Try to find <?xml header in first 100 bytes (just in case there're some XML comments).
	findHeaderUntil := len(b)
	if findHeaderUntil > 100 {
		findHeaderUntil = 100
	}
	if !bytes.Contains(b[:findHeaderUntil], []byte("<?xml")) {
		// No header found. Print it out first.
		w.Write([]byte(xml.Header)) //nolint:errcheck
	}

	w.Write(b) //nolint:errcheck
	return nil
}

// NoContent returns a HTTP 204 "No Content" response.
func NoContent(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func channelEventStream(w http.ResponseWriter, r *http.Request, v interface{}) {
	if reflect.TypeOf(v).Kind() != reflect.Chan {
		panic(fmt.Sprintf("render: event stream expects a channel, not %v", reflect.TypeOf(v).Kind()))
	}

	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")

	if r.ProtoMajor == 1 {
		// An endpoint MUST NOT generate an HTTP/2 message containing connection-specific header fields.
		// Source: RFC7540
		w.Header().Set("Connection", "keep-alive")
	}

	w.WriteHeader(http.StatusOK)

	ctx := r.Context()
	for {
		switch chosen, recv, ok := reflect.Select([]reflect.SelectCase{
			{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ctx.Done())},
			{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(v)},
		}); chosen {
		case 0: // equivalent to: case <-ctx.Done()
			w.Write([]byte("event: error\ndata: {\"error\":\"Server Timeout\"}\n\n")) //nolint:errcheck
			return

		default: // equivalent to: case v, ok := <-stream
			if !ok {
				w.Write([]byte("event: EOF\n\n")) //nolint:errcheck
				return
			}
			v := recv.Interface()

			// Build each channel item.
			if rv, ok := v.(Renderer); ok {
				err := renderer(w, r, rv)
				if err != nil {
					v = err
				} else {
					v = rv
				}
			}

			bytes, err := json.Marshal(v)
			if err != nil {
				w.Write([]byte(fmt.Sprintf("event: error\ndata: {\"error\":\"%v\"}\n\n", err))) //nolint:errcheck
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
				continue
			}
			w.Write([]byte(fmt.Sprintf("event: data\ndata: %s\n\n", bytes))) //nolint:errcheck
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

// channelIntoSlice buffers channel data into a slice.
func channelIntoSlice(w http.ResponseWriter, r *http.Request, from interface{}) interface{} {
	ctx := r.Context()

	var to []interface{}
	for {
		switch chosen, recv, ok := reflect.Select([]reflect.SelectCase{
			{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ctx.Done())},
			{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(from)},
		}); chosen {
		case 0: // equivalent to: case <-ctx.Done()
			http.Error(w, "Server Timeout", http.StatusGatewayTimeout)
			return nil

		default: // equivalent to: case v, ok := <-stream
			if !ok {
				return to
			}
			v := recv.Interface()

			// Render each channel item.
			if rv, ok := v.(Renderer); ok {
				err := renderer(w, r, rv)
				if err != nil {
					v = err
				} else {
					v = rv
				}
			}

			to = append(to, v)
		}
	}
}
