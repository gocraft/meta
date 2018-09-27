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
	assertEqual(t, inputs.WithCamelCase.Path, "with_camel_case")
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

type withNullable struct {
	A String
	B String `meta_null:"true"`
	C String `meta_null:"true"`
}

var withNullableDecoder = NewDecoder(withNullable{})

func TestMapSource(t *testing.T) {
	var inputs withNullable
	e := withNullableDecoder.DecodeMap(&inputs, map[string]interface{}{"a": "1", "b": nil})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val, "1")
	assertEqual(t, inputs.B.Present, true)
	assertEqual(t, inputs.B.Null, true)
	assertEqual(t, inputs.C.Present, false)
	assertEqual(t, inputs.C.Null, false)
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
	i64 := Int64{1, Nullity{false}, Presence{true}, ""}
	valuer = i64
	v, err = valuer.Value()
	assert(t, err == nil)
	assertEqual(t, int64(1), v)

	ui64 := Uint64{1, Nullity{false}, Presence{true}, ""}
	valuer = ui64
	v, err = valuer.Value()
	assert(t, err == nil)
	assertEqual(t, int64(1), v)
}

type withMetaStar struct {
	AField    String
	AllFields map[string]interface{} `meta:"*"`
}

type nestedMetaStar struct {
	Nested *withMetaStar
}

var withMetaStarDecoder = NewDecoder(withMetaStar{})
var nestedMetaStarDecoder = NewDecoder(nestedMetaStar{})

func TestMetaStar(t *testing.T) {
	var inputs withMetaStar
	e := withMetaStarDecoder.Decode(&inputs, url.Values{"cf_other_field": {"Another field"}}, []byte(`{"a_field": "A field", "cf_numeric_field": 12}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.AField.Val, "A field")
	assertEqual(t, len(inputs.AllFields), 3)
	assertEqual(t, inputs.AllFields["cf_numeric_field"], json.Number("12"))
	assertEqual(t, inputs.AllFields["cf_other_field"], "Another field")
	assertEqual(t, inputs.AllFields["a_field"], "A field")
}

func TestNestedMetaStar(t *testing.T) {
	var inputs nestedMetaStar
	e := nestedMetaStarDecoder.Decode(&inputs, url.Values{"nested.cf_other_field": {"Another field"}}, []byte(`{"nested":{"a_field": "A field", "cf_numeric_field": 12}}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.Nested.AField.Val, "A field")
	assertEqual(t, len(inputs.Nested.AllFields), 3)
	assertEqual(t, inputs.Nested.AllFields["cf_numeric_field"], json.Number("12"))
	assertEqual(t, inputs.Nested.AllFields["cf_other_field"], "Another field")
	assertEqual(t, inputs.Nested.AllFields["a_field"], "A field")
}

type WithSelfReference struct {
	Name     String
	Children []*WithSelfReference
}

var withSelfReferenceDecoder = NewDecoder(WithSelfReference{})

func TestWithSelfReference(t *testing.T) {
	var inputs WithSelfReference
	e := withSelfReferenceDecoder.Decode(&inputs, nil, []byte(`{"name": "parent", "children": [{"name": "child 1"}, {"name": "child 2", "children": [{"name": "grandchild"}]}]}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.Name.Val, "parent")
	assertEqual(t, len(inputs.Children), 2)
	assertEqual(t, inputs.Children[0].Name.Val, "child 1")
	assertEqual(t, len(inputs.Children[0].Children), 0)
	assertEqual(t, inputs.Children[1].Name.Val, "child 2")
	assertEqual(t, len(inputs.Children[1].Children), 1)
	assertEqual(t, inputs.Children[1].Children[0].Name.Val, "grandchild")
	assertEqual(t, len(inputs.Children[1].Children[0].Children), 0)
	assertEqual(t, inputs.Children[1].Children[0].Name.Path, "children.1.children.0.name")
}

type ObjectValuer struct {
	Val map[string]interface{}
}

func (v *ObjectValuer) ParseOptions(tag reflect.StructTag) interface{} {
	return nil
}

func (v *ObjectValuer) JSONValue(path string, i interface{}, options interface{}) Errorable {
	v.Val = i.(map[string]interface{})
	return nil
}

type WithObjectValuer struct {
	A ObjectValuer
}

var objectValuerDecoder = NewDecoder(WithObjectValuer{})

func TestObjectValuer(t *testing.T) {
	// json
	var inputs WithObjectValuer
	e := objectValuerDecoder.Decode(&inputs, nil, []byte(`{"a": {"b": 1}}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, len(inputs.A.Val), 1)
	assertEqual(t, json.Number("1"), inputs.A.Val["b"])

	// url.Values
	inputs = WithObjectValuer{}
	e = objectValuerDecoder.DecodeValues(&inputs, url.Values{"a.b": {"1"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, len(inputs.A.Val), 1)
	assertEqual(t, "1", inputs.A.Val["b"])
}

// TODO: test default values
