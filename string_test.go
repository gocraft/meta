package meta

import (
	"net/url"
	"testing"
	"unicode/utf8"
)

type withString struct {
	A String `meta_required:"true"`
}

var withStringDecoder = NewDecoder(&withString{})

func TestStringSuccess(t *testing.T) {
	var inputs withString

	e := withStringDecoder.DecodeValues(&inputs, url.Values{"a": {"Ok"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, "Ok")
	assertEqual(t, inputs.A.Present, true)

	e = withStringDecoder.DecodeJSON(&inputs, []byte(`{"a":"Okay"}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, "Okay")
	assertEqual(t, inputs.A.Present, true)
}

func TestStringRequired(t *testing.T) { // Missing -> required
	var inputs withString

	e := withStringDecoder.DecodeValues(&inputs, url.Values{"b": {"Ok"}})
	assertEqual(t, e, ErrorHash{"a": ErrorAtom("required")})
	assertEqual(t, inputs.A.Val, "")
	assertEqual(t, inputs.A.Present, false)

	e = withStringDecoder.DecodeJSON(&inputs, []byte(`{"b": "ok"}`))
	assertEqual(t, e, ErrorHash{"a": ErrorAtom("required")})
	assertEqual(t, inputs.A.Val, "")
	assertEqual(t, inputs.A.Present, false)
}

func TestStringBlank(t *testing.T) { // "" -> blank
	var inputs withString

	e := withStringDecoder.DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
	assertEqual(t, inputs.A.Val, "")
	assertEqual(t, inputs.A.Present, false)

	e = withStringDecoder.DecodeJSON(&inputs, []byte(`{"a": ""}`))
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
	assertEqual(t, inputs.A.Val, "")
	assertEqual(t, inputs.A.Present, false)

	e = withStringDecoder.DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
	assertEqual(t, inputs.A.Val, "")
	assertEqual(t, inputs.A.Present, false)
}

func TestStringBlank2(t *testing.T) { // "" -> blank
	var inputs struct {
		A String `meta_required:"true" meta_blank:"false"`
	}

	e := NewDecoder(&inputs).DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
	assertEqual(t, inputs.A.Val, "")
	assertEqual(t, inputs.A.Present, false)

	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a": ""}`))
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
	assertEqual(t, inputs.A.Val, "")
	assertEqual(t, inputs.A.Present, false)

	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
	assertEqual(t, inputs.A.Val, "")
	assertEqual(t, inputs.A.Present, false)
}

func TestStringBlankSuccess(t *testing.T) { // "" and `meta_blank:"true"` -> Success
	var inputs struct {
		A String `meta_required:"true" meta_blank:"true"`
	}

	e := NewDecoder(&inputs).DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, "")
	assertEqual(t, inputs.A.Present, true)

	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a": ""}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, "")
	assertEqual(t, inputs.A.Present, true)

	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, "")
	assertEqual(t, inputs.A.Present, true)
}

func TestStringMinRunes(t *testing.T) {
	var inputs struct {
		A String `meta_required:"true" meta_min_runes:"2"`
	}
	d := NewDecoder(&inputs)

	e := d.DecodeValues(&inputs, url.Values{"a": {"ab"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, "ab")

	e = d.DecodeJSON(&inputs, []byte(`{"a":"cd"}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, "cd")

	e = d.DecodeValues(&inputs, url.Values{"a": {"abcdefg"}})
	assertEqual(t, e, ErrorHash(nil))

	e = d.DecodeJSON(&inputs, []byte(`{"a":"12345"}`))
	assertEqual(t, e, ErrorHash(nil))

	e = d.DecodeValues(&inputs, url.Values{"a": {"a"}})
	assertEqual(t, e, ErrorHash{"a": ErrMinRunes})

	e = d.DecodeJSON(&inputs, []byte(`{"a":"a"}`))
	assertEqual(t, e, ErrorHash{"a": ErrMinRunes})

	e = d.DecodeValues(&inputs, url.Values{"a": {"世"}}) // 3-byte character. 1 rune.
	assertEqual(t, e, ErrorHash{"a": ErrMinRunes})

	e = d.DecodeJSON(&inputs, []byte(`{"a":"世"}`))
	assertEqual(t, e, ErrorHash{"a": ErrMinRunes})

	// This involves coersion and then string check
	e = d.DecodeJSON(&inputs, []byte(`{"a":1}`))
	assertEqual(t, e, ErrorHash{"a": ErrMinRunes})
}

func TestStringMaxRunes(t *testing.T) {
	var inputs struct {
		A String `meta_required:"true" meta_max_runes:"2"`
	}
	d := NewDecoder(&inputs)

	e := d.DecodeValues(&inputs, url.Values{"a": {"ab"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, "ab")

	e = d.DecodeJSON(&inputs, []byte(`{"a":"cd"}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, "cd")

	e = d.DecodeValues(&inputs, url.Values{"a": {"a"}})
	assertEqual(t, e, ErrorHash(nil))

	e = d.DecodeJSON(&inputs, []byte(`{"a":"a"}`))
	assertEqual(t, e, ErrorHash(nil))

	e = d.DecodeValues(&inputs, url.Values{"a": {"abc"}})
	assertEqual(t, e, ErrorHash{"a": ErrMaxRunes})

	e = d.DecodeJSON(&inputs, []byte(`{"a":"ade"}`))
	assertEqual(t, e, ErrorHash{"a": ErrMaxRunes})

	e = d.DecodeValues(&inputs, url.Values{"a": {"世"}}) // 3-byte character. 1 rune.
	assertEqual(t, e, ErrorHash(nil))

	e = d.DecodeJSON(&inputs, []byte(`{"a":"世"}`))
	assertEqual(t, e, ErrorHash(nil))

	// This involves coersion and then string check
	e = d.DecodeJSON(&inputs, []byte(`{"a":true}`))
	assertEqual(t, e, ErrorHash{"a": ErrMaxRunes})
}

func TestStringMaxBytes(t *testing.T) {
	var inputs struct {
		A String `meta_required:"true" meta_max_bytes:"2"`
	}
	d := NewDecoder(&inputs)

	e := d.DecodeValues(&inputs, url.Values{"a": {"ab"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, "ab")

	e = d.DecodeJSON(&inputs, []byte(`{"a":"cd"}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, "cd")

	e = d.DecodeValues(&inputs, url.Values{"a": {"a"}})
	assertEqual(t, e, ErrorHash(nil))

	e = d.DecodeJSON(&inputs, []byte(`{"a":"a"}`))
	assertEqual(t, e, ErrorHash(nil))

	e = d.DecodeValues(&inputs, url.Values{"a": {"abc"}})
	assertEqual(t, e, ErrorHash{"a": ErrMaxBytes})

	e = d.DecodeJSON(&inputs, []byte(`{"a":"ade"}`))
	assertEqual(t, e, ErrorHash{"a": ErrMaxBytes})

	e = d.DecodeValues(&inputs, url.Values{"a": {"世"}}) // 3-byte character. 1 rune.
	assertEqual(t, e, ErrorHash{"a": ErrMaxBytes})

	e = d.DecodeJSON(&inputs, []byte(`{"a":"世"}`))
	assertEqual(t, e, ErrorHash{"a": ErrMaxBytes})

	// This involves coersion and then string check
	e = d.DecodeJSON(&inputs, []byte(`{"a":true}`))
	assertEqual(t, e, ErrorHash{"a": ErrMaxBytes})
}

func TestStringIn(t *testing.T) {
	var inputs struct {
		A String `meta_required:"true" meta_in:"foo,bar,baz,true"`
	}
	d := NewDecoder(&inputs)

	e := d.DecodeValues(&inputs, url.Values{"a": {"foo"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, "foo")

	e = d.DecodeJSON(&inputs, []byte(`{"a":"bar"}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, "bar")

	e = d.DecodeValues(&inputs, url.Values{"a": {"bar"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, "bar")

	e = d.DecodeJSON(&inputs, []byte(`{"a":"foo"}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, "foo")

	e = d.DecodeValues(&inputs, url.Values{"a": {"omg"}})
	assertEqual(t, e, ErrorHash{"a": ErrIn})

	e = d.DecodeJSON(&inputs, []byte(`{"a":"omg"}`))
	assertEqual(t, e, ErrorHash{"a": ErrIn})

	e = d.DecodeJSON(&inputs, []byte(`{"a":"true"}`))
	assertEqual(t, e, ErrorHash(nil))

	// Conversion happens first
	e = d.DecodeJSON(&inputs, []byte(`{"a":true}`))
	assertEqual(t, e, ErrorHash(nil))
}

func TestStringStrip(t *testing.T) {
	var inputs withString

	e := withStringDecoder.DecodeValues(&inputs, url.Values{"a": {" ab \n"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, "ab")

	e = withStringDecoder.DecodeJSON(&inputs, []byte(`{"a":" cd "}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, "cd")

	e = withStringDecoder.DecodeValues(&inputs, url.Values{"a": {"  "}})
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
	assertEqual(t, inputs.A.Present, true)

	e = withStringDecoder.DecodeJSON(&inputs, []byte(`{"a":"   "}`))
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
	assertEqual(t, inputs.A.Present, true)
}

func TestStringUtf8(t *testing.T) {
	var inputs withString

	e := withStringDecoder.DecodeValues(&inputs, url.Values{"a": {string([]byte{0x34, 0xFF, 0xFE})}})
	assertEqual(t, e, ErrorHash{"a": ErrUtf8})

	// JSON utf8:
	e = withStringDecoder.DecodeJSON(&inputs, []byte{123, 34, 97, 34, 58, 34, 0xFF, 34, 125})
	assertEqual(t, e, ErrorHash(nil)) // no error-- json.Unmarshal will convert invalid UTF-8 to valid UTF8
	assertEqual(t, true, utf8.ValidString(inputs.A.Val))
}

func TestStringJSONCoersion(t *testing.T) {
	var inputs withString

	e := withStringDecoder.DecodeJSON(&inputs, []byte(`{"a":true}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, "true")

	e = withStringDecoder.DecodeJSON(&inputs, []byte(`{"a":false}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, "false")

	e = withStringDecoder.DecodeJSON(&inputs, []byte(`{"a":3.14}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, "3.14")

	e = withStringDecoder.DecodeJSON(&inputs, []byte(`{"a":["wat"]}`))
	assertEqual(t, e, ErrorHash{"a": ErrString})

	e = withStringDecoder.DecodeJSON(&inputs, []byte(`{"a":{"wat":"ok"}}`))
	assertEqual(t, e, ErrorHash{"a": ErrString})
}

//
// OptionalString
//

type withOptionalString struct {
	A String
}

var withOptionalStringDecoder = NewDecoder(&withOptionalString{})

func TestOptionalStringSuccess(t *testing.T) {
	var inputs withOptionalString
	e := withOptionalStringDecoder.DecodeValues(&inputs, url.Values{"a": {"Ok"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Val, "Ok")

	e = withOptionalStringDecoder.DecodeJSON(&inputs, []byte(`{"a":"Okay"}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Val, "Okay")
}

func TestOptionalStringOmitted(t *testing.T) {
	var inputs withOptionalString

	e := withOptionalStringDecoder.DecodeValues(&inputs, url.Values{"b": {"Ok"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val, "")

	e = withOptionalStringDecoder.DecodeJSON(&inputs, []byte(`{"b":"Okay"}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val, "")
}

func TestOptionalStringBlankOmitted(t *testing.T) {
	var inputs withOptionalString

	e := withOptionalStringDecoder.DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val, "")

	e = withOptionalStringDecoder.DecodeValues(&inputs, url.Values{"a": {"   \n"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val, "")

	e = withOptionalStringDecoder.DecodeJSON(&inputs, []byte(`{"a":""}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val, "")

	e = withOptionalStringDecoder.DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val, "")

	e = withOptionalStringDecoder.DecodeJSON(&inputs, []byte(`{"a":"  "}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val, "")
}

func TestOptionalStringBlank(t *testing.T) {
	var inputs struct {
		A String `meta_discard_blank:"false"`
	}

	e := NewDecoder(&inputs).DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Val, "")

	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a":""}`))
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Val, "")

	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Val, "")

	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a":" "}`))
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Val, "")
}

func TestOptionalStringBlankSuccess(t *testing.T) {
	var inputs struct {
		A String `meta_discard_blank:"false" meta_blank:"true"`
	}

	e := NewDecoder(&inputs).DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Val, "")

	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a":""}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Val, "")

	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Val, "")
}

type withStringPointers struct {
	A *String `meta_required:"true"`
	B *String
}

var withStringPointersDecoder = NewDecoder(&withStringPointers{})

func TestStringPointerSuccess(t *testing.T) {
	var inputs withStringPointers

	e := withStringPointersDecoder.DecodeValues(&inputs, url.Values{"a": {"Ok"}})
	assertEqual(t, e, ErrorHash(nil))
	assert(t, inputs.A != nil)
	assertEqual(t, inputs.A.Val, "Ok")
	assertEqual(t, inputs.B, (*String)(nil))

	e = withStringPointersDecoder.DecodeJSON(&inputs, []byte(`{"a":"Okay"}`))
	assertEqual(t, e, ErrorHash(nil))
	assert(t, inputs.A != nil)
	assertEqual(t, inputs.A.Val, "Okay")
	assertEqual(t, inputs.B, (*String)(nil))
}

func TestStringPointerRequired(t *testing.T) { // Missing -> required
	var inputs withStringPointers

	e := withStringPointersDecoder.DecodeValues(&inputs, url.Values{"f": {"Ok"}})
	assertEqual(t, e, ErrorHash{"a": ErrRequired})
	assertEqual(t, inputs.A, (*String)(nil))
	assertEqual(t, inputs.B, (*String)(nil))

	e = withStringPointersDecoder.DecodeJSON(&inputs, []byte(`{"f":"wat"}`))
	assertEqual(t, e, ErrorHash{"a": ErrRequired})
	assertEqual(t, inputs.A, (*String)(nil))
	assertEqual(t, inputs.B, (*String)(nil))
}

type withOptionalNullString struct {
	A String `meta_null:"true"`
}

var withOptionalNullStringDecoder = NewDecoder(&withOptionalNullString{})

func TestOptionalNullStringSuccess(t *testing.T) {
	var inputs withOptionalNullString
	e := withOptionalNullStringDecoder.DecodeValues(&inputs, url.Values{"a": {"wat"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, false)
	assertEqual(t, inputs.A.Val, "wat")

	inputs = withOptionalNullString{}
	e = withOptionalNullStringDecoder.DecodeJSON(&inputs, []byte(`{"a":"wat"}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, false)
	assertEqual(t, inputs.A.Val, "wat")
}

func TestOptionalNullStringNull(t *testing.T) {
	var inputs withOptionalNullString
	e := withOptionalNullStringDecoder.DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, true)
	assertEqual(t, inputs.A.Val, "")

	inputs = withOptionalNullString{}
	e = withOptionalNullStringDecoder.DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, true)
	assertEqual(t, inputs.A.Val, "")
}

func TestOptionalNullStringOmitted(t *testing.T) {
	var inputs withOptionalNullString
	e := withOptionalNullStringDecoder.DecodeValues(&inputs, url.Values{})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Null, false)
	assertEqual(t, inputs.A.Val, "")

	inputs = withOptionalNullString{}
	e = withOptionalNullStringDecoder.DecodeJSON(&inputs, []byte(`{}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Null, false)
	assertEqual(t, inputs.A.Val, "")
}

type withOptionalNullBlankString struct {
	A String `meta_null:"true" meta_blank:"true"`
}

var withOptionalNullBlankStringDecoder = NewDecoder(&withOptionalNullBlankString{})

func TestOptionalNullBlankStringSuccess(t *testing.T) {
	var inputs withOptionalNullBlankString
	e := withOptionalNullBlankStringDecoder.DecodeValues(&inputs, url.Values{"a": {"wat"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, false)
	assertEqual(t, inputs.A.Val, "wat")

	inputs = withOptionalNullBlankString{}
	e = withOptionalNullBlankStringDecoder.DecodeJSON(&inputs, []byte(`{"a":"wat"}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, false)
	assertEqual(t, inputs.A.Val, "wat")
}

func TestOptionalNullBlankStringNull(t *testing.T) {
	var inputs withOptionalNullBlankString
	e := withOptionalNullBlankStringDecoder.DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, false)
	assertEqual(t, inputs.A.Val, "")

	inputs = withOptionalNullBlankString{}
	e = withOptionalNullBlankStringDecoder.DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, true)
	assertEqual(t, inputs.A.Val, "")

	inputs = withOptionalNullBlankString{}
	e = withOptionalNullBlankStringDecoder.DecodeJSON(&inputs, []byte(`{"a":""}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, false)
	assertEqual(t, inputs.A.Val, "")
}

func TestOptionalNullBlankStringOmitted(t *testing.T) {
	var inputs withOptionalNullBlankString
	e := withOptionalNullBlankStringDecoder.DecodeValues(&inputs, url.Values{})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Null, false)
	assertEqual(t, inputs.A.Val, "")

	inputs = withOptionalNullBlankString{}
	e = withOptionalNullBlankStringDecoder.DecodeJSON(&inputs, []byte(`{}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Null, false)
	assertEqual(t, inputs.A.Val, "")
}
