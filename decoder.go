package render

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"sync"
)

// Decode is a package-level variable set to our default Decoder. We do this
// because it allows you to set render.Decode to another function with the
// same function signature, while also utilizing the render.Decoder() function
// itself. Effectively, allowing you to easily add your own logic to the package
// defaults. For example, maybe you want to impose a limit on the number of
// bytes allowed to be read from the request body.
var Decode = DefaultDecoder

type DecodeFunc func(r io.Reader, v interface{}) error

var decodeMapperLck sync.RWMutex

// decodeMapper will map the generic content type to a decoder
var decodeMapper = map[ContentType]DecodeFunc{
	ContentTypeJSON: DecodeJSON,
	ContentTypeXML:  DecodeXML,
}

// SetDecoder will set the decoder for the given content type.
// Use a nil DecodeFunc to unset a content type
func SetDecoder(contentType ContentType, decoder DecodeFunc) {
	decodeMapperLck.Lock()
	defer decodeMapperLck.Unlock()
	decodeMapper[contentType] = decoder
}

// SupportedDecoders returns a ContentTypeSet of the configured Content types with decoders
func SupportedDecoders() *ContentTypeSet {
	strs := make([]string, 0, len(decodeMapper))
	decodeMapperLck.RLock()
	defer decodeMapperLck.RUnlock()
	for str := range decodeMapper {
		strs = append(strs, string(str))
	}
	sort.Strings(strs)
	return NewContentTypeSet(strs...)
}

func DefaultDecoder(r *http.Request, v interface{}) error {
	decodeMapperLck.RLock()
	defer decodeMapperLck.RUnlock()

	ct := GetRequestContentType(r, ContentTypeNone)

	if decoder := decodeMapper[ct]; decoder != nil {
		return decoder(r.Body, v)
	}
	return fmt.Errorf("render: unable to automatically decode the request content type: '%s'", ct)
}

func DecodeJSON(r io.Reader, v interface{}) error {
	defer io.Copy(ioutil.Discard, r)
	return json.NewDecoder(r).Decode(v)
}

func DecodeXML(r io.Reader, v interface{}) error {
	defer io.Copy(ioutil.Discard, r)
	return xml.NewDecoder(r).Decode(v)
}
