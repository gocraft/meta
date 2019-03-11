package meta

import (
	"fmt"
	"net/url"
	"testing"
)

type withBool struct {
	A Bool `meta_required:"true"`
}

var withBoolDecoder = NewDecoder(&withBool{})

func TestBoolSuccess(t *testing.T) {
	trues := []string{"t", "T", "True", "TRUE", "1", "true"}
	falses := []string{"f", "F", "False", "FALSE", "0", "false"}

	for _, x := range trues {
		var inputs withBool
		e := withBoolDecoder.DecodeValues(&inputs, url.Values{"a": {x}})

		assertEqual(t, e, ErrorHash(nil))
		assertEqual(t, inputs.A.Val, true)
		assertEqual(t, inputs.A.Present, true)
	}

	for _, x := range falses {
		var inputs withBool
		e := withBoolDecoder.DecodeValues(&inputs, url.Values{"a": {x}})

		assertEqual(t, e, ErrorHash(nil))
		assertEqual(t, inputs.A.Val, false)
		assertEqual(t, inputs.A.Present, true)
	}

	jsonTrues := []string{`true`, `"true"`, `1`, `"TRUE"`, `"1"`}
	jsonFalses := []string{`false`, `"false"`, `0`, `"FALSE"`, `"0"`}

	for _, x := range jsonTrues {
		var inputs withBool
		e := withBoolDecoder.DecodeJSON(&inputs, []byte(fmt.Sprintf(`{"a":%s}`, x)))

		assertEqual(t, e, ErrorHash(nil))
		assertEqual(t, inputs.A.Val, true)
		assertEqual(t, inputs.A.Present, true)
	}

	for _, x := range jsonFalses {
		var inputs withBool
		e := withBoolDecoder.DecodeJSON(&inputs, []byte(fmt.Sprintf(`{"a":%s}`, x)))

		assertEqual(t, e, ErrorHash(nil))
		assertEqual(t, inputs.A.Val, false)
		assertEqual(t, inputs.A.Present, true)
	}

	var inputs withBool
	e := withBoolDecoder.DecodeMap(&inputs, map[string]interface{}{"a": true})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, true)
	assertEqual(t, inputs.A.Present, true)
}

func TestBoolBlank(t *testing.T) {
	var inputs withBool
	e := withBoolDecoder.DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
	assertEqual(t, inputs.A.Present, false)

	e = withBoolDecoder.DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
	assertEqual(t, inputs.A.Present, false)

	e = withBoolDecoder.DecodeJSON(&inputs, []byte(`{"a":""}`))
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
	assertEqual(t, inputs.A.Present, false)
}

func TestBoolInvalid(t *testing.T) {
	var inputs withBool
	e := withBoolDecoder.DecodeValues(&inputs, url.Values{"a": {"wat"}})
	assertEqual(t, e, ErrorHash{"a": ErrBool})
	assertEqual(t, inputs.A.Present, false)

	e = withBoolDecoder.DecodeJSON(&inputs, []byte(`{"a":"wat"}`))
	assertEqual(t, e, ErrorHash{"a": ErrBool})
	assertEqual(t, inputs.A.Present, false)

	e = withBoolDecoder.DecodeJSON(&inputs, []byte(`{"a":[true]}`))
	assertEqual(t, e, ErrorHash{"a": ErrBool})
	assertEqual(t, inputs.A.Present, false)

	e = withBoolDecoder.DecodeJSON(&inputs, []byte(`{"a":{"a":true}}`))
	assertEqual(t, e, ErrorHash{"a": ErrBool})
	assertEqual(t, inputs.A.Present, false)

	e = withBoolDecoder.DecodeJSON(&inputs, []byte(`{"a":2}`))
	assertEqual(t, e, ErrorHash{"a": ErrBool})
	assertEqual(t, inputs.A.Present, false)
}

type withOptionalBool struct {
	A Bool
}

var withOptionalBoolDecoder = NewDecoder(&withOptionalBool{})

func TestOptionalBoolSuccess(t *testing.T) {
	var inputs withOptionalBool
	e := withOptionalBoolDecoder.DecodeValues(&inputs, url.Values{"a": {"1"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Val, true)

	inputs = withOptionalBool{}
	e = withOptionalBoolDecoder.DecodeJSON(&inputs, []byte(`{"a":true}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Val, true)
}

func TestOptionalBoolOmitted(t *testing.T) {
	var inputs withOptionalBool

	e := withOptionalBoolDecoder.DecodeValues(&inputs, url.Values{})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val, false)

	inputs = withOptionalBool{}
	e = withOptionalBoolDecoder.DecodeJSON(&inputs, []byte(`{}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val, false)
}

func TestOptionalBoolBlank(t *testing.T) {
	var inputs withOptionalBool

	e := withOptionalBoolDecoder.DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val, false)

	inputs = withOptionalBool{}
	e = withOptionalBoolDecoder.DecodeJSON(&inputs, []byte(`{"a":""}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val, false)

	inputs = withOptionalBool{}
	e = withOptionalBoolDecoder.DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val, false)

}

func TestOptionalBoolBlankFailure(t *testing.T) {
	var inputs struct {
		A Bool `meta_discard_blank:"false"`
	}

	e := NewDecoder(&inputs).DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash{"a": ErrBlank})

	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a":""}`))
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
}

type withOptionalNullBool struct {
	A Bool `meta_null:"true"`
}

var withOptionalNullBoolDecoder = NewDecoder(&withOptionalNullBool{})

func TestOptionalNullBoolSuccess(t *testing.T) {
	var inputs withOptionalNullBool
	e := withOptionalNullBoolDecoder.DecodeValues(&inputs, url.Values{"a": {"true"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, false)
	assertEqual(t, inputs.A.Val, true)

	inputs = withOptionalNullBool{}
	e = withOptionalNullBoolDecoder.DecodeJSON(&inputs, []byte(`{"a":"true"}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, false)
	assertEqual(t, inputs.A.Val, true)
}

func TestOptionalNullBoolNull(t *testing.T) {
	var inputs withOptionalNullBool
	e := withOptionalNullBoolDecoder.DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, true)
	assertEqual(t, inputs.A.Val, false)

	inputs = withOptionalNullBool{}
	e = withOptionalNullBoolDecoder.DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, true)
	assertEqual(t, inputs.A.Val, false)
}

func TestOptionalNullBoolOmitted(t *testing.T) {
	var inputs withOptionalNullBool
	e := withOptionalNullBoolDecoder.DecodeValues(&inputs, url.Values{})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Null, false)
	assertEqual(t, inputs.A.Val, false)

	inputs = withOptionalNullBool{}
	e = withOptionalNullBoolDecoder.DecodeJSON(&inputs, []byte(`{}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Null, false)
	assertEqual(t, inputs.A.Val, false)
}
