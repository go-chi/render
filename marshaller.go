package render

import (
	"encoding/json"
	"io"
)

var jsonMarshaller Marshaller = jsonDefaultMarshaller{}

type Marshaller interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshall(data []byte, v interface{}) error
	NewEncoder(w io.Writer) Encoder
	Decode(r io.Reader, v interface{}) error
}

type Encoder interface {
	SetEscapeHTML(on bool)
	Encode(v interface{}) error
}

type jsonDefaultMarshaller struct{}

func (j jsonDefaultMarshaller) NewEncoder(w io.Writer) Encoder {
	return json.NewEncoder(w)
}

func (j jsonDefaultMarshaller) Decode(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

func (j jsonDefaultMarshaller) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (j jsonDefaultMarshaller) Unmarshall(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func SetJsonMarshaller(m Marshaller) {
	if m == nil {
		return
	}

	jsonMarshaller = m
}
