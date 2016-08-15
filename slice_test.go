package meta

import (
	"net/url"
	"testing"
)

type withSliceString struct {
	A []String `meta_required:"true"`
	B []*String
}

var withSliceStringDecoder = NewDecoder(&withSliceString{})

func TestSliceStringSuccess(t *testing.T) {
	var inputs withSliceString
	e := withSliceStringDecoder.DecodeValues(&inputs, url.Values{
		"a.0": {"z"},
		"a.1": {"y"},
		"a.2": {"x"},
		"b.0": {"w"},
	})

	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, len(inputs.A), 3)
	assertEqual(t, len(inputs.B), 1)
	if len(inputs.A) == 3 {
		assertEqual(t, inputs.A[0].Val, "z")
		assertEqual(t, inputs.A[1].Val, "y")
		assertEqual(t, inputs.A[2].Val, "x")
	}

	if len(inputs.B) == 1 {
		assertEqual(t, inputs.B[0].Val, "w")
	}
}

func TestSliceStringSuccessMultiSource(t *testing.T) {
	var inputs withSliceString
	e := withSliceStringDecoder.Decode(&inputs, url.Values{
		"a.0": {"z"},
		"a.1": {"y"},
		"a.2": {"x"},
		"b.0": {"w"},
	}, []byte(`{"a":["z1", "y1"]}`))

	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, len(inputs.A), 3)
	assertEqual(t, len(inputs.B), 1)
	if len(inputs.A) == 3 {
		assertEqual(t, inputs.A[0].Val, "z1")
		assertEqual(t, inputs.A[1].Val, "y1")
		assertEqual(t, inputs.A[2].Val, "x")
	}

	if len(inputs.B) == 1 {
		assertEqual(t, inputs.B[0].Val, "w")
	}
}

// by default, if no items are present, then the slice will be set to nil
func TestSliceStringNoItems(t *testing.T) {
	var inputs withSliceString
	e := withSliceStringDecoder.DecodeValues(&inputs, url.Values{})

	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, len(inputs.A), 0)
	assertEqual(t, len(inputs.B), 0)
	assertEqual(t, inputs.A, []String(nil))
	assertEqual(t, inputs.B, []*String(nil))
}

// TODO: make a test where it's a []OptionalString, and pass in blank values. Should they be included in the array?

// errors in an element.
func TestSliceStringItemBlank(t *testing.T) {
	var inputs withSliceString
	e := withSliceStringDecoder.DecodeValues(&inputs, url.Values{
		"a.0": {""},
		"a.1": {"z"},
		"a.2": {""},
		"a.3": {"y"},
	})

	assertEqual(t, e, ErrorHash{"a": ErrorSlice{ErrBlank, nil, ErrBlank, nil}})
}

type withSliceOfHashes struct {
	A []struct {
		A String `meta_required:"true"`
		B String
	}

	B []*struct {
		Z String
	}
}

var withSliceOfHashesDecoder = NewDecoder(&withSliceOfHashes{})

func TestSliceOfHashesSuccess(t *testing.T) {
	var inputs withSliceOfHashes
	e := withSliceOfHashesDecoder.DecodeValues(&inputs, url.Values{
		"a.0.a": {"Z"},
		"a.0.b": {"Y"},
		"a.1.a": {"X"},
		"a.1.b": {"W"},
		"a.2.a": {"V"},
		"b.0.z": {""},
		"b.1.z": {"U"},
	})

	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, len(inputs.A), 3)
	assertEqual(t, len(inputs.B), 2)

	if len(inputs.A) == 3 {
		assertEqual(t, inputs.A[0].A.Val, "Z")
		assertEqual(t, inputs.A[0].B.Val, "Y")
		assertEqual(t, inputs.A[0].B.Present, true)

		assertEqual(t, inputs.A[1].A.Val, "X")
		assertEqual(t, inputs.A[1].B.Val, "W")
		assertEqual(t, inputs.A[1].B.Present, true)

		assertEqual(t, inputs.A[2].A.Val, "V")
		assertEqual(t, inputs.A[2].B.Val, "")
		assertEqual(t, inputs.A[2].B.Present, false)
	}

	if len(inputs.B) == 2 {
		// NOTE: it is conceivable that we'd change the spec:
		// if no value is present in a a hash that is in an array, then it's as if the item wasn't passed. So len(inputs.B) == 1 in this case.
		assertEqual(t, inputs.B[0].Z.Val, "")
		assertEqual(t, inputs.B[0].Z.Present, false)

		assertEqual(t, inputs.B[1].Z.Val, "U")
		assertEqual(t, inputs.B[1].Z.Present, true)
	}
}

func TestSliceOfHashesError(t *testing.T) {
	var inputs withSliceOfHashes
	e := withSliceOfHashesDecoder.DecodeValues(&inputs, url.Values{
		"a.0.a": {""}, // blank
		"a.0.b": {"Y"},
		// a.1.a: required
		"a.1.b": {"W"},
		"a.2.a": {"V"}, // ok
	})

	assertEqual(t, e, ErrorHash{"a": ErrorSlice{ErrorHash{"a": ErrBlank}, ErrorHash{"a": ErrRequired}, nil}})
}

func TestSliceOfHashesLength(t *testing.T) {
	type withRequiredSliceOfHashes struct {
		A []struct {
			Z String
		} `meta_min_length:"2" meta_max_length:"4"`
	}

	withRequiredSliceOfHashesDecoder := NewDecoder(&withRequiredSliceOfHashes{})

	// Valid length
	inputs := withRequiredSliceOfHashes{}
	e := withRequiredSliceOfHashesDecoder.DecodeValues(&inputs, url.Values{
		"a.0.z": {"A"},
		"a.1.z": {"B"},
		"a.2.z": {"C"},
	})
	assertEqual(t, e, ErrorHash(nil))

	inputs = withRequiredSliceOfHashes{}
	e = withRequiredSliceOfHashesDecoder.DecodeJSON(&inputs, []byte(`{
		"a": [
			{"z": "A"},
			{"z": "B"},
			{"z": "C"}
		]
	}`))
	assertEqual(t, e, ErrorHash(nil))

	// Too short
	inputs = withRequiredSliceOfHashes{}
	e = withRequiredSliceOfHashesDecoder.DecodeValues(&inputs, url.Values{
		"a.0.z": {"A"},
	})
	assertEqual(t, e, ErrorHash{"a": ErrMinLength})

	inputs = withRequiredSliceOfHashes{}
	e = withRequiredSliceOfHashesDecoder.DecodeJSON(&inputs, []byte(`{
		"a": [
			{"z": "A"}
		]
	}`))
	assertEqual(t, e, ErrorHash{"a": ErrMinLength})

	inputs = withRequiredSliceOfHashes{}
	e = withRequiredSliceOfHashesDecoder.DecodeJSON(&inputs, []byte(`{
		"a": [
			{"z": "A"},
		]
	}`)) // comma
	assertEqual(t, e, ErrorHash{"error": ErrMalformed})

	// Too short
	inputs = withRequiredSliceOfHashes{}
	e = withRequiredSliceOfHashesDecoder.DecodeValues(&inputs, url.Values{})
	assertEqual(t, e, ErrorHash{"a": ErrMinLength})

	inputs = withRequiredSliceOfHashes{}
	e = withRequiredSliceOfHashesDecoder.DecodeJSON(&inputs, []byte(`{}`))
	assertEqual(t, e, ErrorHash{"a": ErrMinLength})

	// Too long
	inputs = withRequiredSliceOfHashes{}
	e = withRequiredSliceOfHashesDecoder.DecodeValues(&inputs, url.Values{
		"a.0.z": {"A"},
		"a.1.z": {"B"},
		"a.2.z": {"C"},
		"a.3.z": {"D"},
		"a.4.z": {"E"},
	})
	assertEqual(t, e, ErrorHash{"a": ErrMaxLength})

	inputs = withRequiredSliceOfHashes{}
	e = withRequiredSliceOfHashesDecoder.DecodeJSON(&inputs, []byte(`{
		"a": [
			{"z": "A"},
			{"z": "B"},
			{"z": "C"},
			{"z": "D"},
			{"z": "E"}
		]
	}`))
	assertEqual(t, e, ErrorHash{"a": ErrMaxLength})
}
