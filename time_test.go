package meta

import (
	"net/url"
	"testing"
	"time"
)

type withTime struct {
	A Time `meta_required:"true"`
}

var withTimeDecoder = NewDecoder(&withTime{})

func TestTimeSuccess(t *testing.T) {
	var inputs withTime

	e := withTimeDecoder.DecodeValues(&inputs, url.Values{"a": {"2015-06-02T16:33:22Z"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val.Equal(time.Date(2015, 6, 2, 16, 33, 22, 0, time.UTC)), true)
	assertEqual(t, inputs.A.Present, true)

	e = withTimeDecoder.DecodeJSON(&inputs, []byte(`{"a":"2016-06-02T16:33:22Z"}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val.Equal(time.Date(2016, 6, 2, 16, 33, 22, 0, time.UTC)), true)
	assertEqual(t, inputs.A.Present, true)
}

func TestTimeBlank(t *testing.T) {
	var inputs withTime

	e := withTimeDecoder.DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
	assertEqual(t, inputs.A.Present, false)

	e = withTimeDecoder.DecodeJSON(&inputs, []byte(`{"a":""}`))
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
	assertEqual(t, inputs.A.Present, false)
}

func TestTimeInvalid(t *testing.T) {
	var inputs withTime

	e := withTimeDecoder.DecodeValues(&inputs, url.Values{"a": {"wat"}})
	assertEqual(t, e, ErrorHash{"a": ErrTime})
	assertEqual(t, inputs.A.Present, false)

	e = withTimeDecoder.DecodeJSON(&inputs, []byte(`{"a":"ok"}`))
	assertEqual(t, e, ErrorHash{"a": ErrTime})
	assertEqual(t, inputs.A.Present, false)
}

func TestTimeCustomFormat(t *testing.T) {
	var inputs struct {
		A Time `meta_required:"true" meta_format:"1/2/2006"`
	}

	e := NewDecoder(&inputs).DecodeValues(&inputs, url.Values{"a": {"6/2/2015"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val.Equal(time.Date(2015, 6, 2, 0, 0, 0, 0, time.UTC)), true)
	assertEqual(t, inputs.A.Present, true)

	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a":"9/1/2015"}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Val.Equal(time.Date(2015, 9, 1, 0, 0, 0, 0, time.UTC)), true)
	assertEqual(t, inputs.A.Present, true)
}

type withOptionalTime struct {
	A Time
}

var withOptionalTimeDecoder = NewDecoder(&withOptionalTime{})

func TestOptionalTimeSuccess(t *testing.T) {
	var inputs withOptionalTime

	e := withOptionalTimeDecoder.DecodeValues(&inputs, url.Values{"a": {"2015-06-02T16:33:22Z"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Val.Equal(time.Date(2015, 6, 2, 16, 33, 22, 0, time.UTC)), true)

	e = withOptionalTimeDecoder.DecodeJSON(&inputs, []byte(`{"a":"2016-06-02T16:33:22Z"}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Val.Equal(time.Date(2016, 6, 2, 16, 33, 22, 0, time.UTC)), true)
}

func TestOptionalTimeOmitted(t *testing.T) {
	var inputs withOptionalTime

	e := withOptionalTimeDecoder.DecodeValues(&inputs, url.Values{})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val.IsZero(), true)

	e = withOptionalTimeDecoder.DecodeJSON(&inputs, []byte(`{"b":"9/1/2015"}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val.IsZero(), true)
}

func TestOptionalTimeBlank(t *testing.T) {
	var inputs withOptionalTime

	e := withOptionalTimeDecoder.DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val.IsZero(), true)

	e = withOptionalTimeDecoder.DecodeJSON(&inputs, []byte(`{"a":""}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val.IsZero(), true)

	e = withOptionalTimeDecoder.DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Val.IsZero(), true)
}

func TestOptionalTimeBlankFailure(t *testing.T) {
	var inputs struct {
		A Time `meta_discard_blank:"false"`
	}

	e := NewDecoder(&inputs).DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash{"a": ErrBlank})

	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a":""}`))
	assertEqual(t, e, ErrorHash{"a": ErrBlank})

	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
}
