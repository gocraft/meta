package meta

import (
	"net/url"
	"testing"
)

//
// StringSlice
//

type withStringSlice struct {
	A StringSlice `meta_required:"true"`
}

var withStringSliceDecoder = NewDecoder(&withStringSlice{})

func TestStringSliceSuccess(t *testing.T) {
	var inputs, inputs2 withStringSlice

	e := withStringSliceDecoder.DecodeValues(&inputs, url.Values{"a": {"aaa,bbb,ccc"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, []string{"aaa", "bbb", "ccc"})

	e = withStringSliceDecoder.DecodeValues(&inputs2, url.Values{"a": {"aaa"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs2.A.Val, []string{"aaa"})

	// JSON - strings get split on commas
	e = withStringSliceDecoder.DecodeJSON(&inputs, []byte(`{"a":"ccc,bbb,aaa"}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, []string{"ccc", "bbb", "aaa"})

	e = withStringSliceDecoder.DecodeJSON(&inputs, []byte(`{"a":"fff"}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, []string{"fff"})

	// Actual array
	e = withStringSliceDecoder.DecodeJSON(&inputs, []byte(`{"a":["rrr", "kkk"]}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, []string{"rrr", "kkk"})

	// Array with conversion
	e = withStringSliceDecoder.DecodeJSON(&inputs, []byte(`{"a":["rrr", true, false, 1]}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, []string{"rrr", "true", "false", "1"})

	// decode map
	e = withStringSliceDecoder.DecodeMap(&inputs, map[string]interface{}{"a": []interface{}{"aaa", "bbb"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, []string{"aaa", "bbb"})
}

func TestStringSliceBlank(t *testing.T) {
	var inputs withStringSlice

	e := withStringSliceDecoder.DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash{"a": ErrBlank})

	e = withStringSliceDecoder.DecodeJSON(&inputs, []byte(`{"a":[]}`))
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
}

func TestStringSliceBlankEntries(t *testing.T) {
	var inputs withStringSlice

	e := withStringSliceDecoder.DecodeValues(&inputs, url.Values{"a": {"aaa,bbb,,,ccc"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, []string{"aaa", "bbb", "ccc"})

	e = withStringSliceDecoder.DecodeJSON(&inputs, []byte(`{"a":"aaaa,bbbb,,,cccc"}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, []string{"aaaa", "bbbb", "cccc"})

	e = withStringSliceDecoder.DecodeJSON(&inputs, []byte(`{"a":["a","b","",null,"","c"]}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, []string{"a", "b", "c"})
}

func TestStringSliceBlankEntriesDiscardBlankFalse(t *testing.T) {
	var inputs struct {
		A StringSlice `meta_required:"true" meta_discard_blank:"false"`
	}

	e := NewDecoder(&inputs).DecodeValues(&inputs, url.Values{"a": {"aaa,bbb,,,ccc"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, []string{"aaa", "bbb", "", "", "ccc"})

	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a":"aaaa,bbbb,,,cccc"}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, []string{"aaaa", "bbbb", "", "", "cccc"})

	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a":["a","b","",null,"","c"]}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, []string{"a", "b", "", "", "", "c"})
}

func TestStringSliceBlankEntriesDiscardBlankFalseBlankFalse(t *testing.T) {
	var inputs struct {
		A StringSlice `meta_required:"true" meta_discard_blank:"false" meta_blank:"false"`
	}

	e := NewDecoder(&inputs).DecodeValues(&inputs, url.Values{"a": {"aaa,bbb,,,ccc"}})
	assertEqual(t, e, ErrorHash{"a": ErrorSlice{nil, nil, ErrBlank, ErrBlank, nil}})
	assertEqual(t, inputs.A.Val, []string{"aaa", "bbb", "", "", "ccc"})

	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a":["a","b","",null,"","c"]}`))
	assertEqual(t, e, ErrorHash{"a": ErrorSlice{nil, nil, ErrBlank, ErrBlank, ErrBlank, nil}})
	assertEqual(t, inputs.A.Val, []string{"a", "b", "", "", "", "c"})
}

func TestStringSliceStrip(t *testing.T) {
	var inputs withStringSlice

	e := withStringSliceDecoder.DecodeValues(&inputs, url.Values{"a": {" wat , who"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, []string{"wat", "who"})

	e = withStringSliceDecoder.DecodeJSON(&inputs, []byte(`{"a":[" ok "]}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, []string{"ok"})
}

func TestStringSliceStripFalse(t *testing.T) {
	var inputs struct {
		A StringSlice `meta_required:"true" meta_strip:"false"`
	}

	e := NewDecoder(&inputs).DecodeValues(&inputs, url.Values{"a": {" wat , who"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, []string{" wat ", " who"})

	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a":[" wut "," whom"]}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, []string{" wut ", " whom"})
}

func TestStringSliceLength(t *testing.T) {
	type boundLengthInput struct {
		A StringSlice `meta_required:"true" meta_min_length:"2" meta_max_length:"4"`
	}

	// Valid length
	inputs := boundLengthInput{}
	e := NewDecoder(&inputs).DecodeValues(&inputs, url.Values{"a": {"aaa,bbb,ccc"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, []string{"aaa", "bbb", "ccc"})

	inputs = boundLengthInput{}
	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a": ["aaa", "bbb", "ccc"]}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, []string{"aaa", "bbb", "ccc"})

	// Too short
	inputs = boundLengthInput{}
	e = NewDecoder(&inputs).DecodeValues(&inputs, url.Values{"a": {"aaa"}})
	assertEqual(t, e, ErrorHash{"a": ErrMinLength})
	assertEqual(t, len(inputs.A.Val), 0)

	inputs = boundLengthInput{}
	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a": ["aaa"]}`))
	assertEqual(t, e, ErrorHash{"a": ErrMinLength})
	assertEqual(t, len(inputs.A.Val), 0)

	// Too long
	inputs = boundLengthInput{}
	e = NewDecoder(&inputs).DecodeValues(&inputs, url.Values{"a": {"aaa,bbb,ccc,ddd,eee"}})
	assertEqual(t, e, ErrorHash{"a": ErrMaxLength})
	assertEqual(t, len(inputs.A.Val), 0)

	inputs = boundLengthInput{}
	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a": ["aaa","bbb","ccc","ddd","eee"]}`))
	assertEqual(t, e, ErrorHash{"a": ErrMaxLength})
	assertEqual(t, len(inputs.A.Val), 0)

}
