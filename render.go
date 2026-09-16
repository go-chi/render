package render

import (
	"net/http"
	"reflect"
)

// Renderer interface for managing response payloads.
type Renderer interface {
	Render(w http.ResponseWriter, r *http.Request) error
}

// Binder interface for managing request payloads.
type Binder interface {
	Bind(r *http.Request) error
}

// Bind decodes a request body and executes the Binder method of the
// payload structure.
func Bind(r *http.Request, v Binder) error {
	if err := Decode(r, v); err != nil {
		return err
	}
	return binder(r, v)
}

// Render renders a single payload and respond to the client request.
func Render(w http.ResponseWriter, r *http.Request, v Renderer) error {
	if err := renderer(w, r, v); err != nil {
		return err
	}
	Respond(w, r, v)
	return nil
}

// RenderList renders a slice of payloads and responds to the client request.
func RenderList(w http.ResponseWriter, r *http.Request, l []Renderer) error {
	for _, v := range l {
		if err := renderer(w, r, v); err != nil {
			return err
		}
	}
	Respond(w, r, l)
	return nil
}

func isNil(f reflect.Value) bool {
	switch f.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return f.IsNil()
	default:
		return false
	}
}

// Executed top-down
func renderer(w http.ResponseWriter, r *http.Request, v Renderer) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	// We call it top-down.
	if err := v.Render(w, r); err != nil {
		return err
	}

	// We're done if the Renderer isn't a struct object
	if rv.Kind() != reflect.Struct {
		return nil
	}

	// For structs, we call Render on each field that implements Renderer
	for i := 0; i < rv.NumField(); i++ {
		f := rv.Field(i)
		if f.Type().Implements(rendererType) {

			if isNil(f) {
				continue
			}

			fv := f.Interface().(Renderer)
			if err := renderer(w, r, fv); err != nil {
				return err
			}

		}
	}

	return nil
}

func asBinder(f reflect.Value) (Binder, bool) {
	if !f.IsValid() {
		return nil, false
	}
	switch f.Kind() {
	case reflect.Ptr, reflect.Interface:
		if f.IsNil() {
			return nil, false
		}
	}
	if f.Type().Implements(binderType) {
		if !f.CanInterface() {
			return nil, false
		}
		b, ok := f.Interface().(Binder)
		return b, ok
	}
	if f.CanAddr() && f.Addr().Type().Implements(binderType) {
		b, ok := f.Addr().Interface().(Binder)
		return b, ok
	}
	if f.CanInterface() && f.Kind() != reflect.Ptr && f.Kind() != reflect.Interface {
		ptr := reflect.New(f.Type())
		ptr.Elem().Set(f)
		if ptr.Type().Implements(binderType) {
			b, ok := ptr.Interface().(Binder)
			return b, ok
		}
	}
	return nil, false
}

func bindValue(r *http.Request, f reflect.Value) error {
	if !f.IsValid() {
		return nil
	}
	if b, ok := asBinder(f); ok {
		return binder(r, b)
	}
	switch f.Kind() {
	case reflect.Ptr, reflect.Interface:
		if f.IsNil() {
			return nil
		}
		return bindValue(r, f.Elem())
	case reflect.Slice, reflect.Array:
		for i := 0; i < f.Len(); i++ {
			if err := bindValue(r, f.Index(i)); err != nil {
				return err
			}
		}
	case reflect.Map:
		for _, key := range f.MapKeys() {
			if err := bindValue(r, f.MapIndex(key)); err != nil {
				return err
			}
		}
	}
	return nil
}

// Executed bottom-up
func binder(r *http.Request, v Binder) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return v.Bind(r)
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Struct:
		for i := 0; i < rv.NumField(); i++ {
			if err := bindValue(r, rv.Field(i)); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			if err := bindValue(r, rv.Index(i)); err != nil {
				return err
			}
		}
	case reflect.Map:
		for _, key := range rv.MapKeys() {
			if err := bindValue(r, rv.MapIndex(key)); err != nil {
				return err
			}
		}
	}

	return v.Bind(r)
}

var (
	rendererType = reflect.TypeOf(new(Renderer)).Elem()
	binderType   = reflect.TypeOf(new(Binder)).Elem()
)

// contextKey is a value for use with context.WithValue. It's used as
// a pointer so it fits in an interface{} without allocation. This technique
// for defining context keys was copied from Go 1.7's new use of context in net/http.
type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "chi render context value " + k.name
}
