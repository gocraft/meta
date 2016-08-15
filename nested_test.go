package meta

import (
	"net/url"
	"testing"
)

type nested struct {
	A String `meta_required:"true"`
	B String
	C struct {
		D String `meta_required:"true"`
		E String
	} `meta_required:"true"`
	F struct {
		G String `meta_required:"true"`
		H String
	}
}

var withNestedDecoder = NewDecoder(&nested{})

func TestNestedSuccess(t *testing.T) {
	var inputs nested
	e := withNestedDecoder.DecodeValues(&inputs, url.Values{
		"a":   {"1"},
		"c.d": {"2"},
	})

	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, "1")
	assertEqual(t, inputs.C.D.Val, "2")
	assertEqual(t, inputs.C.E.Present, false)
	assertEqual(t, inputs.C.E.Val, "")

	assertEqual(t, inputs.F.G.Val, "")
	assertEqual(t, inputs.F.H.Present, false)
	assertEqual(t, inputs.F.H.Val, "")

	inputs = nested{}
	e = withNestedDecoder.DecodeJSON(&inputs, []byte(`{"a":1,"c":{"d":2}}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, "1")
	assertEqual(t, inputs.C.D.Val, "2")
	assertEqual(t, inputs.C.E.Present, false)
	assertEqual(t, inputs.C.E.Val, "")

	assertEqual(t, inputs.F.G.Val, "")
	assertEqual(t, inputs.F.H.Present, false)
	assertEqual(t, inputs.F.H.Val, "")

	// The downside of this design is that it's not clear whether F.G has a valid value in it.
	// (Use a pointer to a struct if you need to know)
}

func TestNestedErrors(t *testing.T) {
	var inputs nested
	e := withNestedDecoder.DecodeValues(&inputs, url.Values{
		"b": {"1"},
	})
	assertEqual(t, e, ErrorHash{"a": ErrRequired, "c": ErrRequired})

	inputs = nested{}
	e = withNestedDecoder.DecodeJSON(&inputs, []byte(`{"b":1}`))
	assertEqual(t, e, ErrorHash{"a": ErrRequired, "c": ErrRequired})

	inputs = nested{}
	e = withNestedDecoder.DecodeValues(&inputs, url.Values{
		"c.e": {"1"},
	})
	assertEqual(t, e, ErrorHash{"a": ErrRequired, "c": ErrorHash{"d": ErrRequired}})

	inputs = nested{}
	e = withNestedDecoder.DecodeJSON(&inputs, []byte(`{"c":{"e":1}}`))
	assertEqual(t, e, ErrorHash{"a": ErrRequired, "c": ErrorHash{"d": ErrRequired}})
}

type nestedWithPointers struct {
	A String `meta_required:"true"`
	B String
	C *struct {
		D String `meta_required:"true"`
		E String
	} `meta_required:"true"`
	F *struct {
		G String `meta_required:"true"`
		H String
	}
}

var withNestedPointersDecoder = NewDecoder(&nestedWithPointers{})

func TestNestedWithPointersSuccess(t *testing.T) {
	var inputs nestedWithPointers
	e := withNestedPointersDecoder.DecodeValues(&inputs, url.Values{
		"a":   {"1"},
		"c.d": {"2"},
	})

	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, "1")
	assert(t, inputs.C != nil)
	if inputs.C != nil {
		assertEqual(t, inputs.C.D.Val, "2")
		assertEqual(t, inputs.C.E.Present, false)
		assertEqual(t, inputs.C.E.Val, "")
	}

	assert(t, inputs.F == nil)

	inputs = nestedWithPointers{}
	e = withNestedPointersDecoder.DecodeJSON(&inputs, []byte(`{"a":1,"c":{"d":2}}`))

	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, "1")
	assert(t, inputs.C != nil)
	if inputs.C != nil {
		assertEqual(t, inputs.C.D.Val, "2")
		assertEqual(t, inputs.C.E.Present, false)
		assertEqual(t, inputs.C.E.Val, "")
	}

	assert(t, inputs.F == nil)
}

func TestNestedWithPointersErrors(t *testing.T) {
	var inputs nestedWithPointers
	e := withNestedPointersDecoder.DecodeValues(&inputs, url.Values{
		"b": {"1"},
	})
	assertEqual(t, e, ErrorHash{"a": ErrRequired, "c": ErrRequired})

	inputs = nestedWithPointers{}
	e = withNestedPointersDecoder.DecodeJSON(&inputs, []byte(`{"b":1}`))
	assertEqual(t, e, ErrorHash{"a": ErrRequired, "c": ErrRequired})

	inputs = nestedWithPointers{}
	e = withNestedPointersDecoder.DecodeValues(&inputs, url.Values{
		"c.e": {"1"},
	})
	assertEqual(t, e, ErrorHash{"a": ErrRequired, "c": ErrorHash{"d": ErrRequired}})

	inputs = nestedWithPointers{}
	e = withNestedPointersDecoder.DecodeJSON(&inputs, []byte(`{"c":{"e":1}}`))
	assertEqual(t, e, ErrorHash{"a": ErrRequired, "c": ErrorHash{"d": ErrRequired}})
}

type Embedded struct {
	A String `meta_required:"true"`
	B String `meta_required:"true"`
}

type embedder struct {
	Embedded `meta_required:"true"`

	C String `meta_required:"true"`
}

var withEmbedDecoder = NewDecoder(&embedder{})

func TestEmbeddingSuccess(t *testing.T) {
	var inputs embedder
	e := withEmbedDecoder.DecodeValues(&inputs, url.Values{
		"a": {"1"},
		"b": {"2"},
		"c": {"3"},
	})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, "1")
	assertEqual(t, inputs.B.Val, "2")
	assertEqual(t, inputs.C.Val, "3")

	inputs = embedder{}
	e = withEmbedDecoder.DecodeJSON(&inputs, []byte(`{"a":1,"b":2,"c":3}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, "1")
	assertEqual(t, inputs.B.Val, "2")
	assertEqual(t, inputs.C.Val, "3")
}

func TestEmbeddingErrors(t *testing.T) {
	var inputs embedder
	e := withEmbedDecoder.DecodeValues(&inputs, url.Values{
		"a": {"1"},
	})
	assertEqual(t, e, ErrorHash{"b": ErrRequired, "c": ErrRequired})

	inputs = embedder{}
	e = withEmbedDecoder.DecodeJSON(&inputs, []byte(`{"a":1}`))
	assertEqual(t, e, ErrorHash{"b": ErrRequired, "c": ErrRequired})
}

// TODO: Test embedding where embedded struct is a ptr
// TODO: test embedding which embeds even more shit
