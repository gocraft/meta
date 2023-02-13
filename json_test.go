package meta

import (
	"encoding/json"
	"testing"
)

type jsonObj struct {
	Str String
	I   Int64
	UI  Uint64
	F   Float64
	B   Bool
}

func TestJsonUnmarshal(t *testing.T) {

	obj := jsonObj{
		Str: NewString("hello world"),
		I:   NewInt64(11),
		UI:  NewUint64(12),
		F:   NewFloat64(13),
		B:   NewBool(true),
	}

	bs, err := json.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}

	var decode jsonObj
	err = json.Unmarshal(bs, &decode)
	if err != nil {
		t.Fatal(err)
	}

	assertEqual(t, decode.Str.Val, "hello world")
	assertEqual(t, decode.I.Val, int64(11))
	assertEqual(t, decode.UI.Val, uint64(12))
	assertEqual(t, decode.F.Val, float64(13))
	assertEqual(t, decode.B.Val, true)
}
