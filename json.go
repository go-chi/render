package render

import (
	jsonstd "encoding/json"
	"io"

	"github.com/json-iterator/go"
)

var json jsonDecodeEncode

func init() {
	json = new(jsonStd)
}

// RegisterJSONIterator if want to use `github.com/json-iterator/go`, please call RegisterJSONIterator first
func RegisterJSONIterator() {
	json = new(jsonIterator)
}

type jsonDecodeEncode interface {
	NewEncoder(w io.Writer) *jsonEncoder
	NewDecoder(r io.Reader) *jsonDecoder
	Marshal(v interface{}) ([]byte, error)
}

type jsonDecoder struct {
	*jsonstd.Decoder
	iteratorDecoder *jsoniter.Decoder
}

type jsonEncoder struct {
	*jsonstd.Encoder
	iteratorEncoder *jsoniter.Encoder
}

type jsonStd struct {
}

func (j jsonStd) NewEncoder(w io.Writer) *jsonEncoder {
	return &jsonEncoder{
		jsonstd.NewEncoder(w),
		nil,
	}
}

func (j jsonStd) NewDecoder(r io.Reader) *jsonDecoder {
	return &jsonDecoder{
		jsonstd.NewDecoder(r),
		nil,
	}
}

func (j jsonStd) Marshal(v interface{}) ([]byte, error) {
	return jsonstd.Marshal(v)
}

type jsonIterator struct {
}

func (j jsonIterator) NewEncoder(w io.Writer) *jsonEncoder {
	return &jsonEncoder{
		nil,
		jsoniter.NewEncoder(w),
	}
}

func (j jsonIterator) NewDecoder(r io.Reader) *jsonDecoder {
	return &jsonDecoder{
		nil,
		jsoniter.NewDecoder(r),
	}
}

func (j jsonIterator) Marshal(v interface{}) ([]byte, error) {
	return jsoniter.Marshal(v)
}
