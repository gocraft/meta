package meta

import (
	"fmt"
	"math"
	"net/url"
	"testing"
)

//
// Int
//

type withInt struct {
	A Int64 `meta_required:"true"`
	B Int64 `meta_required:"true"`
}

var withIntDecoder = NewDecoder(&withInt{})

func TestIntSuccess(t *testing.T) {
	var inputs withInt
	e := withIntDecoder.DecodeValues(&inputs, url.Values{"a": {"-1"}, "b": {"2"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, int64(-1))
	assertEqual(t, inputs.B.Val, int64(2))

	inputs = withInt{}
	e = withIntDecoder.DecodeJSON(&inputs, []byte(`{"a":-1,"b":2}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, int64(-1))
	assertEqual(t, inputs.B.Val, int64(2))
}

func TestIntBlank(t *testing.T) {
	var inputs withInt
	e := withIntDecoder.DecodeValues(&inputs, url.Values{"a": {""}, "b": {""}})
	assertEqual(t, e, ErrorHash{"a": ErrBlank, "b": ErrBlank})

	inputs = withInt{}
	e = withIntDecoder.DecodeJSON(&inputs, []byte(`{"a":null,"b":null}`))
	assertEqual(t, e, ErrorHash{"a": ErrBlank, "b": ErrBlank})
}

func TestIntInvalid(t *testing.T) {
	var inputs withInt
	e := withIntDecoder.DecodeValues(&inputs, url.Values{"a": {"a"}, "b": {"a"}})
	assertEqual(t, e, ErrorHash{"a": ErrInt, "b": ErrInt})

	inputs = withInt{}
	e = withIntDecoder.DecodeJSON(&inputs, []byte(`{"a":"a","b":"b"}`))
	assertEqual(t, e, ErrorHash{"a": ErrInt, "b": ErrInt})
}

func TestIntRange(t *testing.T) {
	var inputs withInt
	inValues := url.Values{
		"a": {fmt.Sprint(math.MaxInt64)},
		"b": {fmt.Sprint(math.MinInt64)},
	}
	e := withIntDecoder.DecodeValues(&inputs, inValues)
	assertEqual(t, e, ErrorHash(nil))

	inputs = withInt{}
	e = withIntDecoder.DecodeJSON(&inputs, []byte(fmt.Sprintf(`{"a":%d,"b":%d}`, math.MaxInt64, math.MinInt64)))
	assertEqual(t, e, ErrorHash(nil))

	const (
		exMaxInt64 = "9223372036854775808"
		exMinInt64 = "-9223372036854775809"
	)
	outValues := url.Values{
		"a": {exMaxInt64}, // note: maxInt + 1 wraps
		"b": {exMinInt64},
	}
	e = withIntDecoder.DecodeValues(&inputs, outValues)
	assertEqual(t, e, ErrorHash{"a": ErrIntRange, "b": ErrIntRange})

	inputs = withInt{}
	e = withIntDecoder.DecodeJSON(&inputs, []byte(fmt.Sprintf(`{"a":%s,"b":%s}`, exMaxInt64, exMinInt64)))
	assertEqual(t, e, ErrorHash{"a": ErrIntRange, "b": ErrIntRange})
}

type withMinMaxInt struct {
	A Int64 `meta_required:"true" meta_min:"-5" meta_max:"11"`
	B Int64 `meta_required:"true" meta_min:"-5" meta_max:"11"`
	C Int64 `meta_required:"true" meta_min:"-5" meta_max:"11"`
}

var withMinMaxIntDecoder = NewDecoder(&withMinMaxInt{})

func TestIntMinMax(t *testing.T) {
	var inputs withMinMaxInt
	e := withMinMaxIntDecoder.DecodeValues(&inputs, url.Values{"a": {"-5"}, "b": {"6"}, "c": {"11"}})
	assertEqual(t, e, ErrorHash(nil))

	inputs = withMinMaxInt{}
	e = withMinMaxIntDecoder.DecodeJSON(&inputs, []byte(`{"a":-5,"b":6,"c":11}`))
	assertEqual(t, e, ErrorHash(nil))

	inputs = withMinMaxInt{}
	e = withMinMaxIntDecoder.DecodeValues(&inputs, url.Values{"a": {"-6"}, "b": {"16"}, "c": {"6"}})
	assertEqual(t, e, ErrorHash{"a": ErrMin, "b": ErrMax})

	inputs = withMinMaxInt{}
	e = withMinMaxIntDecoder.DecodeJSON(&inputs, []byte(`{"a":-6,"b":16,"c":6}`))
	assertEqual(t, e, ErrorHash{"a": ErrMin, "b": ErrMax})
}

type withInInt struct {
	A Int64 `meta_in:"-4,3,9"`
	B Int64 `meta_in:"-4,3,9"`
	C Int64 `meta_in:"-4,3,9"`
}

var withInIntDecoder = NewDecoder(&withInInt{})

func TestIntIn(t *testing.T) {
	var inputs withInInt
	e := withInIntDecoder.DecodeValues(&inputs, url.Values{"a": {"-4"}, "b": {"3"}, "c": {"9"}})
	assertEqual(t, e, ErrorHash(nil))

	inputs = withInInt{}
	e = withInIntDecoder.DecodeJSON(&inputs, []byte(`{"a":-4,"b":3,"c":9}`))
	assertEqual(t, e, ErrorHash(nil))

	inputs = withInInt{}
	e = withInIntDecoder.DecodeValues(&inputs, url.Values{"a": {"-6"}, "b": {"4"}, "c": {"11"}})
	assertEqual(t, e, ErrorHash{"a": ErrIn, "b": ErrIn, "c": ErrIn})

	inputs = withInInt{}
	e = withInIntDecoder.DecodeJSON(&inputs, []byte(`{"a":-6,"b":4,"c":11}`))
	assertEqual(t, e, ErrorHash{"a": ErrIn, "b": ErrIn, "c": ErrIn})
}

//
// Uint
//

type withUint struct {
	A Uint64 `meta_required:"true"`
	B Uint64 `meta_required:"true"`
}

var withUintDecoder = NewDecoder(&withUint{})

func TestUintSuccess(t *testing.T) {
	var inputs withUint
	e := withUintDecoder.DecodeValues(&inputs, url.Values{"a": {"0"}, "b": {"2"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, uint64(0))
	assertEqual(t, inputs.B.Val, uint64(2))

	inputs = withUint{}
	e = withUintDecoder.DecodeJSON(&inputs, []byte(`{"a":0,"b":2}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, uint64(0))
	assertEqual(t, inputs.B.Val, uint64(2))
}

func TestUintBlank(t *testing.T) {
	var inputs withUint
	e := withUintDecoder.DecodeValues(&inputs, url.Values{"a": {""}, "b": {""}})
	assertEqual(t, e, ErrorHash{"a": ErrBlank, "b": ErrBlank})

	inputs = withUint{}
	e = withUintDecoder.DecodeJSON(&inputs, []byte(`{"a":null,"b":null}`))
	assertEqual(t, e, ErrorHash{"a": ErrBlank, "b": ErrBlank})
}

func TestUintInvalid(t *testing.T) {
	var inputs withUint
	e := withUintDecoder.DecodeValues(&inputs, url.Values{"a": {"a"}, "b": {"a"}})
	assertEqual(t, e, ErrorHash{"a": ErrInt, "b": ErrInt})

	inputs = withUint{}
	e = withUintDecoder.DecodeJSON(&inputs, []byte(`{"a":"a","b":"b"}`))
	assertEqual(t, e, ErrorHash{"a": ErrInt, "b": ErrInt})
}

func TestUintRange(t *testing.T) {
	var inputs withUint
	inValues := url.Values{
		"a": {fmt.Sprint(uint64(0))},
		"b": {fmt.Sprint(uint64(math.MaxUint64))},
	}
	e := withUintDecoder.DecodeValues(&inputs, inValues)
	assertEqual(t, e, ErrorHash(nil))

	inputs = withUint{}
	e = withUintDecoder.DecodeJSON(&inputs, []byte(fmt.Sprintf(`{"a":0,"b":%d}`, uint64(math.MaxUint64))))
	assertEqual(t, e, ErrorHash(nil))

	const (
		exMaxUint64 = "18446744073709551617"
		exMinUint64 = "-1"
	)
	outValues := url.Values{
		"a": {exMaxUint64},
		"b": {exMinUint64},
	}
	e = withUintDecoder.DecodeValues(&inputs, outValues)
	// NOTE: apparently strconv.ParseUint treats "-1" as a syntax error instead of a range error.
	// I think it should be a range error, but I don't currently care as no use cases exists yet.
	assertEqual(t, e, ErrorHash{"a": ErrIntRange, "b": ErrInt})

	inputs = withUint{}
	e = withUintDecoder.DecodeJSON(&inputs, []byte(fmt.Sprintf(`{"a":%s,"b":%s}`, exMaxUint64, exMinUint64)))
	assertEqual(t, e, ErrorHash{"a": ErrIntRange, "b": ErrInt})
}

type withMinMaxUint struct {
	A Uint64 `meta_required:"true" meta_min:"5" meta_max:"11"`
	B Uint64 `meta_required:"true" meta_min:"5" meta_max:"11"`
	C Uint64 `meta_required:"true" meta_min:"5" meta_max:"11"`
	D Uint64 `meta_required:"true" meta_min:"5" meta_max:"11"`
	E Uint64 `meta_required:"true" meta_min:"5" meta_max:"11"`
}

var withMinMaxUintDecoder = NewDecoder(&withMinMaxUint{})

func TestUintMinMax(t *testing.T) {
	var inputs withMinMaxUint
	e := withMinMaxUintDecoder.DecodeValues(&inputs, url.Values{"a": {"5"}, "b": {"11"}, "c": {"6"}, "d": {"1"}, "e": {"16"}})
	assertEqual(t, e, ErrorHash{"d": ErrMin, "e": ErrMax})

	inputs = withMinMaxUint{}
	e = withMinMaxUintDecoder.DecodeJSON(&inputs, []byte(`{"a":5,"b":11,"c":6,"d":1,"e":16}`))
	assertEqual(t, e, ErrorHash{"d": ErrMin, "e": ErrMax})
}

type withInUint struct {
	A Uint64 `meta_required:"true" meta_in:"4,3,9"`
	B Uint64 `meta_required:"true" meta_in:"4,3,9"`
	C Uint64 `meta_required:"true" meta_in:"4,3,9"`
}

var withInUintDecoder = NewDecoder(&withInUint{})

func TestUintIn(t *testing.T) {
	var inputs withInUint
	e := withInUintDecoder.DecodeValues(&inputs, url.Values{"a": {"4"}, "b": {"3"}, "c": {"9"}})
	assertEqual(t, e, ErrorHash(nil))

	inputs = withInUint{}
	e = withInUintDecoder.DecodeJSON(&inputs, []byte(`{"a":4,"b":3,"c":9}`))
	assertEqual(t, e, ErrorHash(nil))

	inputs = withInUint{}
	e = withInUintDecoder.DecodeValues(&inputs, url.Values{"a": {"6"}, "b": {"0"}, "c": {"9"}})
	assertEqual(t, e, ErrorHash{"a": ErrIn, "b": ErrIn})

	inputs = withInUint{}
	e = withInUintDecoder.DecodeJSON(&inputs, []byte(`{"a":6,"b":0,"c":9}`))
	assertEqual(t, e, ErrorHash{"a": ErrIn, "b": ErrIn})
}

type withOptionalInt struct {
	A Int64
}

var withOptionalIntDecoder = NewDecoder(&withOptionalInt{})

func TestOptionalIntSuccess(t *testing.T) {
	var inputs withOptionalInt
	e := withOptionalIntDecoder.DecodeValues(&inputs, url.Values{"a": {"5"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Val, int64(5))

	inputs = withOptionalInt{}
	e = withOptionalIntDecoder.DecodeJSON(&inputs, []byte(`{"a":5}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Val, int64(5))
}

func TestOptionalIntOmitted(t *testing.T) {
	var inputs withOptionalInt
	e := withOptionalIntDecoder.DecodeValues(&inputs, url.Values{})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val, int64(0))

	inputs = withOptionalInt{}
	e = withOptionalIntDecoder.DecodeJSON(&inputs, []byte(`{}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val, int64(0))
}

func TestOptionalIntBlank(t *testing.T) {
	var inputs withOptionalInt
	e := withOptionalIntDecoder.DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val, int64(0))

	inputs = withOptionalInt{}
	e = withOptionalIntDecoder.DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val, int64(0))
}

type withOptionalNonBlankInt struct {
	A Int64 `meta_discard_blank:"false"`
}

var withOptionalNonBlankIntDecoder = NewDecoder(&withOptionalNonBlankInt{})

func TestOptionalIntBlankFailure(t *testing.T) {
	var inputs withOptionalNonBlankInt
	e := withOptionalNonBlankIntDecoder.DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash{"a": ErrBlank})

	inputs = withOptionalNonBlankInt{}
	e = withOptionalNonBlankIntDecoder.DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
}

type withOptionalUint struct {
	A Uint64
}

var withOptionalUintDecoder = NewDecoder(&withOptionalUint{})

func TestOptionalUintSuccess(t *testing.T) {
	var inputs withOptionalUint
	e := withOptionalUintDecoder.DecodeValues(&inputs, url.Values{"a": {"1"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Val, uint64(1))

	inputs = withOptionalUint{}
	e = withOptionalUintDecoder.DecodeJSON(&inputs, []byte(`{"a":1}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Val, uint64(1))
}

func TestOptionalUintOmitted(t *testing.T) {
	var inputs withOptionalUint
	e := withOptionalUintDecoder.DecodeValues(&inputs, url.Values{})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val, uint64(0))

	inputs = withOptionalUint{}
	e = withOptionalUintDecoder.DecodeJSON(&inputs, []byte(`{}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val, uint64(0))
}

func TestOptionalUintBlank(t *testing.T) {
	var inputs withOptionalUint
	e := withOptionalUintDecoder.DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val, uint64(0))

	inputs = withOptionalUint{}
	e = withOptionalUintDecoder.DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val, uint64(0))
}

type withOptionalNonBlankUint struct {
	A Uint64 `meta_discard_blank:"false"`
}

var withOptionalNonBlankUintDecoder = NewDecoder(&withOptionalNonBlankUint{})

func TestOptionalUintBlankFailure(t *testing.T) {
	var inputs withOptionalNonBlankUint
	e := withOptionalNonBlankUintDecoder.DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash{"a": ErrBlank})

	inputs = withOptionalNonBlankUint{}
	e = withOptionalNonBlankUintDecoder.DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
}

//
// OptionalNullInt64
//

type withOptionalNullInt struct {
	A Int64 `meta_null:"true"`
}

var withOptionalNullIntDecoder = NewDecoder(&withOptionalNullInt{})

func TestOptionalNullIntSuccess(t *testing.T) {
	var inputs withOptionalNullInt
	e := withOptionalNullIntDecoder.DecodeValues(&inputs, url.Values{"a": {"5"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, false)
	assertEqual(t, inputs.A.Val, int64(5))

	inputs = withOptionalNullInt{}
	e = withOptionalNullIntDecoder.DecodeJSON(&inputs, []byte(`{"a":5}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, false)
	assertEqual(t, inputs.A.Val, int64(5))
}

func TestOptionalNullIntNull(t *testing.T) {
	var inputs withOptionalNullInt
	e := withOptionalNullIntDecoder.DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, true)
	assertEqual(t, inputs.A.Val, int64(0))

	inputs = withOptionalNullInt{}
	e = withOptionalNullIntDecoder.DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, true)
	assertEqual(t, inputs.A.Val, int64(0))
}

func TestOptionalNullIntOmitted(t *testing.T) {
	var inputs withOptionalNullInt
	e := withOptionalNullIntDecoder.DecodeValues(&inputs, url.Values{})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Null, false)
	assertEqual(t, inputs.A.Val, int64(0))

	inputs = withOptionalNullInt{}
	e = withOptionalNullIntDecoder.DecodeJSON(&inputs, []byte(`{}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Null, false)
	assertEqual(t, inputs.A.Val, int64(0))
}

type withOptionalNullUint struct {
	A Uint64 `meta_null:"true"`
}

var withOptionalNullUintDecoder = NewDecoder(&withOptionalNullUint{})

func TestOptionalNullUintSuccess(t *testing.T) {
	var inputs withOptionalNullUint
	e := withOptionalNullUintDecoder.DecodeValues(&inputs, url.Values{"a": {"5"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, false)
	assertEqual(t, inputs.A.Val, uint64(5))

	inputs = withOptionalNullUint{}
	e = withOptionalNullUintDecoder.DecodeJSON(&inputs, []byte(`{"a":5}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, false)
	assertEqual(t, inputs.A.Val, uint64(5))
}

func TestOptionalNullUintNull(t *testing.T) {
	var inputs withOptionalNullUint
	e := withOptionalNullUintDecoder.DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, true)
	assertEqual(t, inputs.A.Val, uint64(0))

	inputs = withOptionalNullUint{}
	e = withOptionalNullUintDecoder.DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, true)
	assertEqual(t, inputs.A.Val, uint64(0))
}

func TestOptionalNullUintOmitted(t *testing.T) {
	var inputs withOptionalNullUint
	e := withOptionalNullUintDecoder.DecodeValues(&inputs, url.Values{})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Null, false)
	assertEqual(t, inputs.A.Val, uint64(0))

	inputs = withOptionalNullUint{}
	e = withOptionalNullUintDecoder.DecodeJSON(&inputs, []byte(`{}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Null, false)
	assertEqual(t, inputs.A.Val, uint64(0))
}
