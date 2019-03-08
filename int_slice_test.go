package meta

import (
	"net/url"
	"testing"
)

//
// Int
//

type withIntSlice struct {
	A Int64Slice
}

var withIntSliceDecoder = NewDecoder(&withIntSlice{})

func TestIntSliceSuccess(t *testing.T) {
	var inputs, inputs2 withIntSlice

	e := withIntSliceDecoder.DecodeValues(&inputs, url.Values{"a": {"-1,8,3"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, []int64{-1, 8, 3})

	e = withIntSliceDecoder.DecodeValues(&inputs2, url.Values{"a": {"5"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs2.A.Val, []int64{5})

	e = withIntSliceDecoder.DecodeJSON(&inputs, []byte(`{"a":"-2,9,30"}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, []int64{-2, 9, 30})

	e = withIntSliceDecoder.DecodeJSON(&inputs, []byte(`{"a":"6"}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, []int64{6})

	e = withIntSliceDecoder.DecodeJSON(&inputs, []byte(`{"a":[-2,9,"30"]}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, []int64{-2, 9, 30})

	e = withIntSliceDecoder.DecodeMap(&inputs, map[string]interface{}{"a": []interface{}{-2, 9}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, []int64{-2, 9})
}

func TestIntSliceBlank(t *testing.T) {
	var inputs withIntSlice

	e := withIntSliceDecoder.DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash{"a": ErrBlank})

	e = withIntSliceDecoder.DecodeJSON(&inputs, []byte(`{"a":""}`))
	assertEqual(t, e, ErrorHash{"a": ErrBlank})

	e = withIntSliceDecoder.DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash{"a": ErrBlank})

	e = withIntSliceDecoder.DecodeJSON(&inputs, []byte(`{"a":[]}`))
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
}

func TestIntSliceInvalid(t *testing.T) {
	var inputs withIntSlice

	e := withIntSliceDecoder.DecodeValues(&inputs, url.Values{"a": {"1,b"}})
	assertEqual(t, e, ErrorHash{"a": ErrorSlice{nil, ErrInt}})

	e = withIntSliceDecoder.DecodeJSON(&inputs, []byte(`{"a":[1, true, false, {}, [0], "9", "bob"]}`))
	assertEqual(t, e, ErrorHash{"a": ErrorSlice{nil, ErrInt, ErrInt, ErrInt, ErrInt, nil, ErrInt}})
}

func TestIntSliceMinMax(t *testing.T) {
	var inputs struct {
		A Int64Slice `meta_min:"-5" meta_max:"11"`
	}

	e := NewDecoder(&inputs).DecodeValues(&inputs, url.Values{"a": {"-6,4,17"}})
	assertEqual(t, e, ErrorHash{"a": ErrorSlice{ErrMin, nil, ErrMax}})

	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a":[-1, -6, "-10", 12]}`))
	assertEqual(t, e, ErrorHash{"a": ErrorSlice{nil, ErrMin, ErrMin, ErrMax}})
}

func TestIntSliceLength(t *testing.T) {
	type boundLengthInput struct {
		A Int64Slice `meta_min_length:"2" meta_max_length:"4"`
	}

	// Valid length
	inputs := boundLengthInput{}
	e := NewDecoder(&inputs).DecodeValues(&inputs, url.Values{"a": {"0,1,2"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, []int64{0, 1, 2})

	inputs = boundLengthInput{}
	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a": [0, 1, 2]}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, []int64{0, 1, 2})

	// Too short
	inputs = boundLengthInput{}
	e = NewDecoder(&inputs).DecodeValues(&inputs, url.Values{"a": {"0"}})
	assertEqual(t, e, ErrorHash{"a": ErrMinLength})
	assertEqual(t, len(inputs.A.Val), 0)

	inputs = boundLengthInput{}
	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a": [0]}`))
	assertEqual(t, e, ErrorHash{"a": ErrMinLength})
	assertEqual(t, len(inputs.A.Val), 0)

	// Too long
	inputs = boundLengthInput{}
	e = NewDecoder(&inputs).DecodeValues(&inputs, url.Values{"a": {"0,1,2,3,4"}})
	assertEqual(t, e, ErrorHash{"a": ErrMaxLength})
	assertEqual(t, len(inputs.A.Val), 0)

	inputs = boundLengthInput{}
	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a": [0,1,2,3,4]}`))
	assertEqual(t, e, ErrorHash{"a": ErrMaxLength})
	assertEqual(t, len(inputs.A.Val), 0)
}
