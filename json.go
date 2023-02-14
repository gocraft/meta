package meta

import (
	"bytes"
	"encoding/json"
)

var MetaJson JsonLoader

func init() {
	MetaJson = &defaultJsonLoader{}
}

// Expose json loader interface that can be implemneted in client code for customization
type JsonLoader interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
	UnmarshalUsingNumber(data []byte, v interface{}) error
}

// Default implementation of Json loader just uses Go's standard json lib
type defaultJsonLoader struct{}

func (j *defaultJsonLoader) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (j *defaultJsonLoader) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (j *defaultJsonLoader) UnmarshalUsingNumber(data []byte, v interface{}) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	return dec.Decode(v)
}
