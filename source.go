package meta

import (
	"bytes"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
)

// source normalizes form value and json.
type source interface {
	Get(key string) source
	// A pointer must be passed.
	// If the value is not present, the pointer will be nil.
	Value(interface{}) Errorable
	Empty() bool
	ValueMap() map[string]interface{}
	// Malformed source must return ErrMalformed when Value is called.
	// It should be set by the parent source.
	// If the parent is malformed, its children must be malformed.
	Malformed() bool
}

type mergedSource []source

func newMergedSource(src ...source) source {
	return mergedSource(src)
}

func (s mergedSource) Get(key string) source {
	var src []source
	for _, m := range s {
		src = append(src, m.Get(key))
	}
	return newMergedSource(src...)
}

func (s mergedSource) Empty() bool {
	// empty if all sources are empty
	for _, m := range s {
		if !m.Empty() {
			return false
		}
	}
	return true
}

func (s mergedSource) Malformed() bool {
	// malformed if any source is malformed
	for _, m := range s {
		if m.Malformed() {
			return true
		}
	}
	return false
}

func (s mergedSource) Value(i interface{}) Errorable {
	// use the first non-empty source
	for _, m := range s {
		if m.Malformed() {
			return ErrMalformed
		}
		if m.Empty() {
			continue
		}
		return m.Value(i)
	}
	return ErrBlank
}

func (s mergedSource) ValueMap() map[string]interface{} {
	out := make(map[string]interface{})
	for _, m := range s {
		values := m.ValueMap()
		for k, v := range values {
			out[k] = v
		}
	}
	return out
}

func newJSONSource(b []byte) source {
	return &jsonSource{
		RawMessage: b,
	}
}

type jsonSource struct {
	json.RawMessage
	malformed bool
}

func (jv *jsonSource) Empty() bool {
	return len(jv.RawMessage) == 0
}

func (jv *jsonSource) Malformed() bool {
	return jv.malformed
}

func (jv *jsonSource) Get(key string) source {
	s := &jsonSource{
		malformed: jv.malformed,
	}
	if len(jv.RawMessage) == 0 {
		return s
	}
	// numeric key implies array
	i, err := strconv.Atoi(key)
	if err == nil {
		var slice []json.RawMessage
		err = json.Unmarshal(jv.RawMessage, &slice)
		if err != nil {
			s.malformed = true
			return s
		}
		if i >= len(slice) {
			return s
		}
		s.RawMessage = slice[i]
		return s
	}
	var m map[string]json.RawMessage
	err = json.Unmarshal(jv.RawMessage, &m)
	if err != nil {
		s.malformed = true
		return s
	}
	raw, ok := m[key]
	if !ok {
		return s
	}
	s.RawMessage = raw
	return s
}

func (jv *jsonSource) Value(i interface{}) Errorable {
	if len(jv.RawMessage) == 0 {
		return ErrBlank
	}

	dec := json.NewDecoder(bytes.NewReader(jv.RawMessage))
	dec.UseNumber()
	err := dec.Decode(i)
	if err != nil {
		return ErrMalformed
	}
	return nil
}

func (jv *jsonSource) ValueMap() map[string]interface{} {
	var out map[string]interface{}
	if err := json.Unmarshal(jv.RawMessage, &out); err != nil {
		return nil
	}
	return out
}

func newFormValueSource(values url.Values) source {
	return &formValueSource{Values: values}
}

type formValueSource struct {
	url.Values
	prefix string
}

func (fv *formValueSource) Empty() bool {
	for key := range fv.Values {
		if key == fv.prefix || strings.HasPrefix(key, fv.prefix+".") {
			return false
		}
	}
	return true
}

func (fv *formValueSource) Get(key string) source {
	if fv.prefix != "" {
		key = fv.prefix + "." + key
	}
	return &formValueSource{
		Values: fv.Values,
		prefix: key,
	}
}

func (fv *formValueSource) Malformed() bool {
	return false
}

func (fv *formValueSource) Value(i interface{}) Errorable {
	value := fv.Values.Get(fv.prefix)
	switch v := i.(type) {
	case *string:
		*v = value
	case *interface{}:
		*v = value
	default:
		return ErrBlank
	}
	return nil
}

func (fv *formValueSource) ValueMap() map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range fv.Values {
		if len(v) == 1 {
			out[k] = v[0]
		} else {
			out[k] = v
		}
	}
	return out
}
