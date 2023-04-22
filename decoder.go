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
	// Decodes a given reader into an interface
	Decode(r io.Reader, req *http.Request, v interface{}) error
}

// Package-level variables for decoding the supported formats. They are set to
// our default implementations. By setting render.Decode{JSON,XML,Form} you can
// customize Decoding (e.g. you might want to configure the JSON-decoder)
var (
	DecoderJSON Decoder = DecodeJSONInter{}
	DecoderXML  Decoder = DecodeXMLInter{}
	DecoderForm Decoder = DecodeFormInter{}
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

type DecodeJSONInter struct{}

// Decodes a given reader into an interface using the json decoder.
func (DecodeJSONInter) Decode(r io.Reader, req *http.Request, v interface{}) error {
	defer io.Copy(ioutil.Discard, r) //nolint:errcheck
	return json.NewDecoder(r).Decode(v)
}

// DecodeJSON decodes a given reader into an interface using the json decoder.
//
// Deprecated: DecoderJSON.Decode() should be used.
func DecodeJSON(r io.Reader, v interface{}) error {
	defer io.Copy(ioutil.Discard, r) //nolint:errcheck
	return json.NewDecoder(r).Decode(v)
}

type DecodeXMLInter struct{}

// Decodes a given reader into an interface using the xml decoder.
func (DecodeXMLInter) Decode(r io.Reader, req *http.Request, v interface{}) error {
	defer io.Copy(ioutil.Discard, r) //nolint:errcheck
	return xml.NewDecoder(r).Decode(v)
}

// DecodeXML decodes a given reader into an interface using the xml decoder.
//
// Deprecated: DecoderXML.Decode() should be used.
func DecodeXML(r io.Reader, v interface{}) error {
	defer io.Copy(ioutil.Discard, r) //nolint:errcheck
	return xml.NewDecoder(r).Decode(v)
}

type DecodeFormInter struct{}

// Decodes a given reader into an interface using the form decoder.
func (DecodeFormInter) Decode(r io.Reader, req *http.Request, v interface{}) error {
	decoder := form.NewDecoder(r) //nolint:errcheck
	return decoder.Decode(v)
}

// DecodeForm decodes a given reader into an interface using the form decoder.
//
// Deprecated: DecoderForm.Decode() should be used.
func DecodeForm(r io.Reader, v interface{}) error {
	decoder := form.NewDecoder(r) //nolint:errcheck
	return decoder.Decode(v)
}
