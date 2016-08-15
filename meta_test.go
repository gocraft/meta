package meta

import (
	"database/sql/driver"
	"encoding/json"
	"net/url"
	"reflect"
	"testing"
)

func assert(t *testing.T, this bool) {
	if !this {
		t.Errorf("Expected true but was false")
	}
}

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

type withMetaName struct {
	WithCamelCase String `meta_required:"true"`
	OtherField    String `meta_required:"true" meta:"poopin"`
	Ignored       String `meta:"-"`
}

// we test ptr to struct in other places, now we test struct here
var withMetaNameDecoder = NewDecoder(withMetaName{})

func TestNewDecoder(t *testing.T) {
	var inputs withMetaName
	e := withMetaNameDecoder.DecodeValues(&inputs, url.Values{
		"with_camel_case": {"1"},
		"poopin":          {"2"},
		"ignored":         {"3"},
	})

	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.WithCamelCase.Val, "1")
	assertEqual(t, inputs.OtherField.Val, "2")
	assertEqual(t, inputs.Ignored.Val, "")

	inputs = withMetaName{}
	e = withMetaNameDecoder.DecodeJSON(&inputs, []byte(`{"with_camel_case":1,"poopin":2,"ignored":3}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.WithCamelCase.Val, "1")
	assertEqual(t, inputs.OtherField.Val, "2")
	assertEqual(t, inputs.Ignored.Val, "")
}

func TestMalformedJSON(t *testing.T) {
	var inputs withMetaName
	e := withMetaNameDecoder.DecodeJSON(&inputs, []byte(`{"with_camel_case":1,"poopin":2,"ignored":3`)) // malformed json
	assertEqual(t, e, ErrorHash{
		"error": ErrMalformed,
	})
}

func TestMultiSource(t *testing.T) {
	var inputs withMetaName
	e := withMetaNameDecoder.Decode(&inputs, url.Values{"poopin": {"2"}}, []byte(`{"with_camel_case":1,"ignored":3}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.WithCamelCase.Val, "1")
	assertEqual(t, inputs.OtherField.Val, "2")
	assertEqual(t, inputs.Ignored.Val, "")
}

func TestErrorsAreJsonable(t *testing.T) {
	var inputs withMetaName
	e := withMetaNameDecoder.DecodeValues(&inputs, url.Values{})
	j, _ := json.Marshal(e)
	// NOTE: I'm not sure if the order is consistient
	assertEqual(t, string(j), `{"poopin":"required","with_camel_case":"required"}`)

	inputs = withMetaName{}
	e = withMetaNameDecoder.DecodeJSON(&inputs, []byte(`{}`))
	j, _ = json.Marshal(e)
	assertEqual(t, string(j), `{"poopin":"required","with_camel_case":"required"}`)
}

// Ensure that all our types implement the Valuer interface for good compatibility with sql libraries.
func TestValuers(t *testing.T) {
	var valuer driver.Valuer // We're going to use the compiler as our unit tester and assign a bunch of shit to this

	// Bool
	var b Bool
	b.Val = true
	b.Present = true
	valuer = b
	v, err := valuer.Value()
	assert(t, err == nil)
	assertEqual(t, true, v)
	b.Present = false
	valuer = b
	v, err = valuer.Value()
	assert(t, err == nil)
	assertEqual(t, nil, v)

	// String
	var s String
	s.Present = true
	valuer = s
	v, err = valuer.Value()
	assert(t, err == nil)
	assertEqual(t, "", v)

	// Int
	i64 := Int64{1, Nullity{false}, Presence{true}}
	valuer = i64
	v, err = valuer.Value()
	assert(t, err == nil)
	assertEqual(t, int64(1), v)

	ui64 := Uint64{1, Presence{true}}
	valuer = ui64
	v, err = valuer.Value()
	assert(t, err == nil)
	assertEqual(t, int64(1), v)
}

// TODO: test default values
