package render

import (
	"bytes"
	"context"
	"encoding"
	"encoding/binary"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"sync"
)

// M is a convenience alias for quickly building a map structure that is going
// out to a responder. Just a short-hand.
type M map[string]interface{}

// ErrCanNotEncodeObject should be returned by RespondFunc if the Responder should
// try a different content type, as we don't know how to respond with this object
var ErrCanNotEncodeObject = errors.New("error can not encode object")

// Respond is a package-level variable set to our default Responder. We do this
// because it allows you to set render.Respond to another function with the
// same function signature, while also utilizing the render.Responder() function
// itself. Effectively, allowing you to easily add your own logic to the package
// defaults. For example, maybe you want to test if v is an error and respond
// differently, or log something before you respond.
var Respond = DefaultResponder

type RespondFunc func(http.ResponseWriter, *http.Request, interface{}) error

type Responder interface {
	Respond(http.ResponseWriter, *http.Request, interface{}) error
}

var respondMapperLck sync.RWMutex

// respondMapper will map the generic content type to a respond
var respondMapper = map[ContentType]RespondFunc{
	ContentTypeDefault:     JSON,
	ContentTypeJSON:        JSON,
	ContentTypeXML:         XML,
	ContentTypeEventStream: channelEventStream,
}

// SetResponder will set the responder for the given content type.
// Use a nil RespondFunc to unset a content type
func SetResponder(contentType ContentType, responder RespondFunc) {
	respondMapperLck.Lock()
	defer respondMapperLck.Unlock()
	respondMapper[contentType] = responder
}

// SupportedResponders returns a ContentTypeSet of the configured Content types with responders
func SupportedResponders() *ContentTypeSet {
	strs := make([]string, 0, len(respondMapper))
	respondMapperLck.RLock()
	defer respondMapperLck.RUnlock()
	for str := range respondMapper {
		strs = append(strs, string(str))
	}
	sort.Strings(strs)
	return NewContentTypeSet(strs...)
}

// StatusCtxKey is a context key to record a future HTTP response status code.
var StatusCtxKey = &contextKey{"Status"}

// Status sets a HTTP response status code hint into request context at any point
// during the request life-cycle. Before the Responder sends its response header
// it will check the StatusCtxKey
func Status(r *http.Request, status int) {
	*r = *r.WithContext(context.WithValue(r.Context(), StatusCtxKey, status))
}

// DefaultResponder handles streaming JSON and XML responses, automatically setting the
// Content-Type based on request headers. It will default to a JSON response.
func DefaultResponder(w http.ResponseWriter, r *http.Request, v interface{}) {
	var err error

	acceptedTypes := GetAcceptedContentType(r)
	if v != nil {
		switch reflect.TypeOf(v).Kind() {
		case reflect.Chan:
			if acceptedTypes.Has(ContentTypeEventStream) {
				respondMapperLck.RLock()
				if fn, ok := respondMapper[ContentTypeEventStream]; ok {
					respondMapperLck.RUnlock()
					if err = fn(w, r, v); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
					return
				}
				respondMapperLck.Unlock()
			}
			v = channelIntoSlice(w, r, v)
		}
	}

	respondMapperLck.RLock()
	defer respondMapperLck.RUnlock()

	for acceptedTypes.Next() {
		// Skip ContentTypeEventStream, handled up top.
		if acceptedTypes.Type() == ContentTypeEventStream {
			continue
		}
		if fn, ok := respondMapper[acceptedTypes.Type()]; ok {
			if err = fn(w, r, v); err != nil {

				if errors.Is(err, ErrCanNotEncodeObject) {
					// Let's try the next content type
					continue
				}

				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}
	if err = respondMapper[ContentTypeDefault](w, r, v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func setNosniff(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
}

// PlainText writes a string to the response, setting the Content-Type as
// text/plain.
func PlainText(w http.ResponseWriter, r *http.Request, v interface{}) error {
	var txt string

	switch vv := v.(type) {
	case encoding.TextMarshaler:
		btxt, err := vv.MarshalText()
		if err != nil {
			return err
		}
		txt = string(btxt)
	case string:
		txt = vv
	case fmt.Stringer:
		txt = vv.String()
	default:
		return ErrCanNotEncodeObject
	}

	setNosniff(w)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if status, ok := r.Context().Value(StatusCtxKey).(int); ok {
		w.WriteHeader(status)
	}
	w.Write([]byte(txt))
	return nil
}

// Data writes raw bytes to the response, setting the Content-Type as
// application/octet-stream.
func Data(w http.ResponseWriter, r *http.Request, v []byte) {
	setNosniff(w)
	w.Header().Set("Content-Type", "application/octet-stream")
	if status, ok := r.Context().Value(StatusCtxKey).(int); ok {
		w.WriteHeader(status)
	}
	var (
		b   []byte
		err error
	)

	switch vv := v.(type) {
	case encoding.BinaryMarshaler:
		b, err = vv.MarshalBinary()
		if err != nil {
			return err
		}
	case []byte:
		b = vv
	case encoding.TextMarshaler:
		t, err := vv.MarshalText()
		if err != nil {
			return err
		}
		b = []byte(t)
	case string:
		b = []byte(vv)
	case fmt.Stringer:
		b = []byte(vv.String())

	default:
		return binary.Write(w, binary.BigEndian, v)
	}
	w.Write(b)
	return nil
}

// HTML writes a string to the response, setting the Content-Type as text/html.
func HTML(w http.ResponseWriter, r *http.Request, v interface{}) error {
	var txt string

	switch vv := v.(type) {
	case encoding.TextMarshaler:
		btxt, err := vv.MarshalText()
		if err != nil {
			return err
		}
		txt = string(btxt)
	case string:
		txt = vv
	case fmt.Stringer:
		txt = vv.String()
	default:
		return ErrCanNotEncodeObject
	}

	setNosniff(w)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if status, ok := r.Context().Value(StatusCtxKey).(int); ok {
		w.WriteHeader(status)
	}
	w.Write([]byte(txt))
	return nil
}

// JSON marshals 'v' to JSON, automatically escaping HTML and setting the
// Content-Type as application/json.
func JSON(w http.ResponseWriter, r *http.Request, v interface{}) error {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("JSON encode: %w", err)
	}
	setNosniff(w)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if status, ok := r.Context().Value(StatusCtxKey).(int); ok {
		w.WriteHeader(status)
	}
	w.Write(buf.Bytes())
	return nil
}

// XML marshals 'v' to XML, setting the Content-Type as application/xml. It
// will automatically prepend a generic XML header (see encoding/xml.Header) if
// one is not found in the first 100 bytes of 'v'.
func XML(w http.ResponseWriter, r *http.Request, v interface{}) error {
	b, err := xml.Marshal(v)
	if err != nil {
		return fmt.Errorf("XML marshal: %w", err)
	}
	setNosniff(w)
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
		w.Write([]byte(xml.Header))
	}

	w.Write(b)
	return nil
}

// NoContent returns a HTTP 204 "No Content" response.
func NoContent(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(204)
}

func channelEventStream(w http.ResponseWriter, r *http.Request, v interface{}) error {
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

	w.WriteHeader(200)

	ctx := r.Context()
	for {
		switch chosen, recv, ok := reflect.Select([]reflect.SelectCase{
			{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ctx.Done())},
			{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(v)},
		}); chosen {
		case 0: // equivalent to: case <-ctx.Done()
			w.Write([]byte("event: error\ndata: {\"error\":\"Server Timeout\"}\n\n"))
			return nil

		default: // equivalent to: case v, ok := <-stream
			if !ok {
				w.Write([]byte("event: EOF\n\n"))
				return nil
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
				w.Write([]byte(fmt.Sprintf("event: error\ndata: {\"error\":\"%v\"}\n\n", err)))
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
				continue
			}
			w.Write([]byte(fmt.Sprintf("event: data\ndata: %s\n\n", bytes)))
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
			http.Error(w, "Server Timeout", 504)
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
