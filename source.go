package meta

import (
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
	// the path to this source from the root
	Path() string
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

func (s mergedSource) Path() string {
	if len(s) > 0 {
		return s[0].Path()
	} else {
		return ""
	}
}

func newJSONSource(b []byte) source {
	return &jsonSource{
		RawMessage: b,
	}
}

type jsonSource struct {
	json.RawMessage
	malformed bool
	path      string
}

func (jv *jsonSource) Empty() bool {
	return len(jv.RawMessage) == 0
}

func (jv *jsonSource) Malformed() bool {
	return jv.malformed
}

func (jv *jsonSource) Get(key string) source {
	var path string
	if jv.path == "" {
		path = key
	} else {
		path = jv.path + "." + key
	}
	s := &jsonSource{
		malformed: jv.malformed,
		path:      path,
	}
	if len(jv.RawMessage) == 0 {
		return s
	}
	// numeric key implies array
	i, err := strconv.Atoi(key)
	if err == nil {
		var slice []json.RawMessage
		err = MetaJson.Unmarshal(jv.RawMessage, &slice)
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
	err = MetaJson.Unmarshal(jv.RawMessage, &m)
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

	err := MetaJson.UnmarshalUsingNumber(jv.RawMessage, i)
	if err != nil {
		return ErrMalformed
	}
	return nil
}

func (jv *jsonSource) ValueMap() map[string]interface{} {
	var out map[string]interface{}

	err := MetaJson.UnmarshalUsingNumber(jv.RawMessage, &out)
	if err != nil {
		return nil
	}
	return out
}

func (jv *jsonSource) Path() string {
	return jv.path
}

func newFormValueSource(urlValues url.Values) source {
	root := make(map[string]interface{})
	for key, v := range urlValues {
		keyParts := strings.Split(key, ".")
		m := root
		for i, k := range keyParts {
			if i == len(keyParts)-1 {
				if len(v) == 1 {
					m[k] = v[0]
				} else {
					m[k] = v
				}
			} else {
				m2, ok := m[k].(map[string]interface{})
				if !ok {
					m2 = make(map[string]interface{})
					m[k] = m2
				}
				m = m2
			}
		}
	}
	if len(root) > 0 {
		return &formValueSource{value: root}
	}
	return &formValueSource{}
}

type formValueSource struct {
	value interface{}
	path  string
}

func (fv *formValueSource) Empty() bool {
	return fv.value == nil
}

func (fv *formValueSource) Get(key string) source {
	path := key
	if fv.path != "" {
		path = fv.path + "." + key
	}
	if m, ok := fv.value.(map[string]interface{}); ok {
		return &formValueSource{value: m[key], path: path}
	}
	return &formValueSource{path: path}
}

func (fv *formValueSource) Malformed() bool {
	return false
}

func (fv *formValueSource) Value(i interface{}) Errorable {
	switch v := i.(type) {
	case *interface{}:
		*v = fv.value
	default:
		return ErrBlank
	}
	return nil
}

func (fv *formValueSource) ValueMap() map[string]interface{} {
	if m, ok := fv.value.(map[string]interface{}); ok {
		return m
	}
	return map[string]interface{}{}
}

func (fv *formValueSource) Path() string {
	return fv.path
}
