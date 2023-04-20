package render

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/ajg/form"
)

// Decoder is a generic interface to decode arbitrary data from a reader `r` to
// a value `v`.
type Decoder interface {
	Decode(r io.Reader, req *http.Request, v interface{}) error
}

// Package-level variables for decoding the supported formats. They are set to
// our default implementations. By setting render.Decode{JSON,XML,Form} you can
// customize Decoding (e.g. you might want to configure the JSON-decoder)
// TODO documentation
var (
	DecoderJSON Decoder = DecodeJSON{}
	DecoderXML  Decoder = DecodeXML{}
	DecoderForm Decoder = DecodeForm{}
)

// Decode is a package-level variable set to our default Decoder. We do this
// because it allows you to set render.Decode to another function with the
// same function signature, while also utilizing the render.Decoder() function
// itself. Effectively, allowing you to easily add your own logic to the package
// defaults. For example, maybe you want to impose a limit on the number of
// bytes allowed to be read from the request body.
var Decode = DefaultDecoder

// DefaultDecoder detects the correct decoder for use on an HTTP request and
// marshals into a given interface.
func DefaultDecoder(r *http.Request, v interface{}) error {
	var err error

	switch GetRequestContentType(r) {
	case ContentTypeJSON:
		err = DecoderJSON.Decode(r.Body, r, v)
	case ContentTypeXML:
		err = DecoderXML.Decode(r.Body, r, v)
	case ContentTypeForm:
		err = DecoderForm.Decode(r.Body, r, v)
	default:
		err = errors.New("render: unable to automatically decode the request content type")
	}

	return err
}

type DecodeJSON struct{}

// DecodeJSON decodes a given reader into an interface using the json decoder.
func (DecodeJSON) Decode(r io.Reader, req *http.Request, v interface{}) error {
	defer io.Copy(ioutil.Discard, r) //nolint:errcheck
	return json.NewDecoder(r).Decode(v)
}

type DecodeXML struct{}

// DecodeXML decodes a given reader into an interface using the xml decoder.
func (DecodeXML) Decode(r io.Reader, req *http.Request, v interface{}) error {
	defer io.Copy(ioutil.Discard, r) //nolint:errcheck
	return xml.NewDecoder(r).Decode(v)
}

type DecodeForm struct{}

// DecodeForm decodes a given reader into an interface using the form decoder.
func (DecodeForm) Decode(r io.Reader, req *http.Request, v interface{}) error {
	decoder := form.NewDecoder(r) //nolint:errcheck
	return decoder.Decode(v)
}
