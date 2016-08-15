package meta

import (
	"fmt"
	"math"
	"net/url"
	"testing"
)

//
// Float
//

type withFloat struct {
	A Float64 `meta_required:"true"`
	B Float64 `meta_required:"true"`
}

var withFloatDecoder = NewDecoder(&withFloat{})

func TestFloatSuccess(t *testing.T) {
	var inputs withFloat

	e := withFloatDecoder.DecodeValues(&inputs, url.Values{"a": {"-1.1"}, "b": {"2.2"}})

	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, -1.1)
	assertEqual(t, inputs.B.Val, 2.2)

	inputs = withFloat{}
	e = withFloatDecoder.DecodeJSON(&inputs, []byte(`{"a":-1.1,"b":2.2}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, -1.1)
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, false)
	assertEqual(t, inputs.B.Val, 2.2)
	assertEqual(t, inputs.B.Present, true)
	assertEqual(t, inputs.B.Null, false)

	inputs = withFloat{}
	e = withFloatDecoder.DecodeJSON(&inputs, []byte(`{"a":-1,"b":2}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, float64(-1))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, false)
	assertEqual(t, inputs.B.Val, float64(2))
	assertEqual(t, inputs.B.Present, true)
	assertEqual(t, inputs.B.Null, false)
}

func TestFloatBlank(t *testing.T) {
	var inputs withFloat
	e := withFloatDecoder.DecodeValues(&inputs, url.Values{"a": {""}, "b": {""}})
	assertEqual(t, e, ErrorHash{"a": ErrBlank, "b": ErrBlank})
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Null, false)

	inputs = withFloat{}
	e = withFloatDecoder.DecodeJSON(&inputs, []byte(`{"a":null,"b":null}`))
	assertEqual(t, e, ErrorHash{"a": ErrBlank, "b": ErrBlank})
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Null, false)
}

func TestFloatInvalid(t *testing.T) {
	var inputs withFloat
	e := withFloatDecoder.DecodeValues(&inputs, url.Values{"a": {"a"}, "b": {"a"}})
	assertEqual(t, e, ErrorHash{"a": ErrFloat, "b": ErrFloat})
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Null, false)

	inputs = withFloat{}
	e = withFloatDecoder.DecodeJSON(&inputs, []byte(`{"a":"a","b":"b"}`))
	assertEqual(t, e, ErrorHash{"a": ErrFloat, "b": ErrFloat})
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Null, false)
}

func TestFloatRange(t *testing.T) {
	var inputs withFloat

	inValues := url.Values{
		"a": {fmt.Sprint(math.MaxFloat64)},
		"b": {"1.0"},
	}
	e := withFloatDecoder.DecodeValues(&inputs, inValues)
	assertEqual(t, e, ErrorHash(nil))

	inputs = withFloat{}
	e = withFloatDecoder.DecodeJSON(&inputs, []byte(fmt.Sprintf(`{"a":%f,"b":%f}`, math.MaxFloat64, 1.0)))
	assertEqual(t, e, ErrorHash(nil))

	const (
		exMaxFloat64 = "2.797693134862315708145274237317043567981e+308"  // a little higher than 2**1023 * (2**53 - 1) / 2**52
		exMinFloat64 = "-2.797693134862315708145274237317043567981e+308" // a little lower than -2**1023 * (2**53 - 1) / 2**52
	)
	outValues := url.Values{
		"a": {exMaxFloat64}, "b": {exMinFloat64},
	}
	e = withFloatDecoder.DecodeValues(&inputs, outValues)
	assertEqual(t, e, ErrorHash{"a": ErrFloatRange, "b": ErrFloatRange})

	inputs = withFloat{}
	e = withFloatDecoder.DecodeJSON(&inputs, []byte(fmt.Sprintf(`{"a":%s,"b":%s}`, exMaxFloat64, exMinFloat64)))
	assertEqual(t, e, ErrorHash{"a": ErrFloatRange, "b": ErrFloatRange})
}

type withMinMaxFloat struct {
	A Float64 `meta_required:"true" meta_min:"-5.0" meta_max:"11.0"`
	B Float64 `meta_required:"true" meta_min:"-5.0" meta_max:"11.0"`
	C Float64 `meta_required:"true" meta_min:"-5.0" meta_max:"11.0"`
}

var withMinMaxFloatDecoder = NewDecoder(&withMinMaxFloat{})

func TestFloatMinMax(t *testing.T) {
	var inputs withMinMaxFloat
	e := withMinMaxFloatDecoder.DecodeValues(&inputs, url.Values{"a": {"-5.000"}, "b": {"6.0001"}, "c": {"11.0"}})
	assertEqual(t, e, ErrorHash(nil))

	inputs = withMinMaxFloat{}
	e = withMinMaxFloatDecoder.DecodeValues(&inputs, url.Values{"a": {"-5.00001"}, "b": {"16.0"}, "c": {"11.00000000"}})
	assertEqual(t, e, ErrorHash{"a": ErrMin, "b": ErrMax})

	inputs = withMinMaxFloat{}
	e = withMinMaxFloatDecoder.DecodeJSON(&inputs, []byte(`{"a":-5.000,"b":6.0001,"c":11.0}`))
	assertEqual(t, e, ErrorHash(nil))

	inputs = withMinMaxFloat{}
	e = withMinMaxFloatDecoder.DecodeJSON(&inputs, []byte(`{"a":-5.00001,"b":16.0,"c":11.00000000}`))
	assertEqual(t, e, ErrorHash{"a": ErrMin, "b": ErrMax})
}

type withInFloat struct {
	A Float64 `meta_required:"true" meta_in:"-4.1,3.0000,9.1"`
	B Float64 `meta_required:"true" meta_in:"-4.1,3.0000,9.1"`
	C Float64 `meta_required:"true" meta_in:"-4.1,3.0000,9.1"`
}

var withInFloatDecoder = NewDecoder(&withInFloat{})

func TestFloatIn(t *testing.T) {
	var inputs withInFloat
	e := withInFloatDecoder.DecodeValues(&inputs, url.Values{"a": {"-4.10"}, "b": {"3.0"}, "c": {"9.1"}})
	assertEqual(t, e, ErrorHash(nil))

	inputs = withInFloat{}
	e = withInFloatDecoder.DecodeValues(&inputs, url.Values{"a": {"-4.1111"}, "b": {"3.1"}, "c": {"-9.1"}})
	assertEqual(t, e, ErrorHash{"a": ErrIn, "b": ErrIn, "c": ErrIn})

	inputs = withInFloat{}
	e = withInFloatDecoder.DecodeJSON(&inputs, []byte(`{"a":-4.10,"b":3.0,"c":9.1}`))
	assertEqual(t, e, ErrorHash(nil))

	inputs = withInFloat{}
	e = withInFloatDecoder.DecodeJSON(&inputs, []byte(`{"a":-4.1111,"b":3.1,"c":-9.1}`))
	assertEqual(t, e, ErrorHash{"a": ErrIn, "b": ErrIn, "c": ErrIn})
}

type withOptionalFloat struct {
	A Float64
}

var withOptionalFloatDecoder = NewDecoder(&withOptionalFloat{})

func TestOptionalFloatSuccess(t *testing.T) {
	var inputs withOptionalFloat
	e := withOptionalFloatDecoder.DecodeValues(&inputs, url.Values{"a": {"4.20"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Val, 4.20)

	inputs = withOptionalFloat{}
	e = withOptionalFloatDecoder.DecodeJSON(&inputs, []byte(`{"a":4.20}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Val, 4.20)
}

func TestOptionalFloatOmitted(t *testing.T) {
	var inputs withOptionalFloat
	e := withOptionalFloatDecoder.DecodeValues(&inputs, url.Values{})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val, float64(0))

	inputs = withOptionalFloat{}
	e = withOptionalFloatDecoder.DecodeJSON(&inputs, []byte(`{}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val, float64(0))
}

func TestOptionalFloatBlank(t *testing.T) {
	var inputs withOptionalFloat
	e := withOptionalFloatDecoder.DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val, float64(0))

	inputs = withOptionalFloat{}
	e = withOptionalFloatDecoder.DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val, float64(0))
}

type withOptionalNonBlankFloat struct {
	A Float64 `meta_discard_blank:"false"`
}

var withOptionalNonBlankFloatDecoder = NewDecoder(&withOptionalNonBlankFloat{})

func TestOptionalFloatBlankFailure(t *testing.T) {
	var inputs withOptionalNonBlankFloat
	e := withOptionalNonBlankFloatDecoder.DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash{"a": ErrBlank})

	inputs = withOptionalNonBlankFloat{}
	e = withOptionalNonBlankFloatDecoder.DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
}

//
// OptionalNullFloat64
//

type withOptionalNullFloat struct {
	A Float64 `meta_null:"true"`
}

var withOptionalNullFloatDecoder = NewDecoder(&withOptionalNullFloat{})

func TestOptionalNullFloatSuccess(t *testing.T) {
	var inputs withOptionalNullFloat
	e := withOptionalNullFloatDecoder.DecodeValues(&inputs, url.Values{"a": {"5.1"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, false)
	assertEqual(t, inputs.A.Val, 5.1)

	inputs = withOptionalNullFloat{}
	e = withOptionalNullFloatDecoder.DecodeJSON(&inputs, []byte(`{"a":5.1}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, false)
	assertEqual(t, inputs.A.Val, 5.1)
}

func TestOptionalNullFloatNull(t *testing.T) {
	var inputs withOptionalNullFloat
	e := withOptionalNullFloatDecoder.DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, true)
	assertEqual(t, inputs.A.Val, float64(0))

	inputs = withOptionalNullFloat{}
	e = withOptionalNullFloatDecoder.DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, true)
	assertEqual(t, inputs.A.Val, float64(0))
}

func TestOptionalNullFloatOmitted(t *testing.T) {
	var inputs withOptionalNullFloat
	e := withOptionalNullFloatDecoder.DecodeValues(&inputs, url.Values{})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Null, false)
	assertEqual(t, inputs.A.Val, float64(0))

	inputs = withOptionalNullFloat{}
	e = withOptionalNullFloatDecoder.DecodeJSON(&inputs, []byte(`{}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Null, false)
	assertEqual(t, inputs.A.Val, float64(0))
}
