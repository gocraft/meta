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
	assert(t, inputs.A.Val.Equal(time.Date(2015, 6, 2, 16, 33, 22, 0, time.UTC)))
	assertEqual(t, inputs.A.Present, true)

	e = withTimeDecoder.DecodeJSON(&inputs, []byte(`{"a":"2016-06-02T16:33:22Z"}`))
	assertEqual(t, e, ErrorHash(nil))
	assert(t, inputs.A.Val.Equal(time.Date(2016, 6, 2, 16, 33, 22, 0, time.UTC)))
	assertEqual(t, inputs.A.Present, true)

	e = withTimeDecoder.DecodeMap(&inputs, map[string]interface{}{"a": time.Date(2016, 6, 2, 16, 33, 22, 0, time.UTC)})
	assertEqual(t, e, ErrorHash(nil))
	assert(t, inputs.A.Val.Equal(time.Date(2016, 6, 2, 16, 33, 22, 0, time.UTC)))
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

	e = withTimeDecoder.DecodeMap(&inputs, map[string]interface{}{"a": time.Time{}})
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
	assert(t, inputs.A.Val.Equal(time.Date(2015, 6, 2, 0, 0, 0, 0, time.UTC)))
	assertEqual(t, inputs.A.Present, true)

	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a":"9/1/2015"}`))
	assertEqual(t, e, ErrorHash(nil))
	assert(t, inputs.A.Val.Equal(time.Date(2015, 9, 1, 0, 0, 0, 0, time.UTC)))
	assertEqual(t, inputs.A.Present, true)
}

func TestMultipleFormats(t *testing.T) {
	var inputs struct {
		A Time
	}

	e := NewDecoderWithOptions(&inputs, DecoderOptions{
		TimeFormats: []string{time.RFC3339, "2006-01-02 15:04:05"},
	}).DecodeValues(&inputs, url.Values{"a": {"2016-01-01 00:00:00"}})
	assertEqual(t, e, ErrorHash(nil))
	assert(t, inputs.A.Val.Equal(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)))
	assertEqual(t, inputs.A.Present, true)

	e = NewDecoderWithOptions(&inputs, DecoderOptions{
		TimeFormats: []string{time.RFC3339, "2006-01-02 15:04:05"},
	}).DecodeValues(&inputs, url.Values{"a": {"2016-01-01T00:00:00Z"}})
	assertEqual(t, e, ErrorHash(nil))
	assert(t, inputs.A.Val.Equal(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)))
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
	assert(t, inputs.A.Val.Equal(time.Date(2015, 6, 2, 16, 33, 22, 0, time.UTC)))

	e = withOptionalTimeDecoder.DecodeJSON(&inputs, []byte(`{"a":"2016-06-02T16:33:22Z"}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assert(t, inputs.A.Val.Equal(time.Date(2016, 6, 2, 16, 33, 22, 0, time.UTC)))
}

func TestOptionalTimeOmitted(t *testing.T) {
	var inputs withOptionalTime

	e := withOptionalTimeDecoder.DecodeValues(&inputs, url.Values{})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assert(t, inputs.A.Val.IsZero())

	e = withOptionalTimeDecoder.DecodeJSON(&inputs, []byte(`{"b":"9/1/2015"}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assert(t, inputs.A.Val.IsZero())
}

func TestOptionalTimeBlank(t *testing.T) {
	var inputs withOptionalTime

	e := withOptionalTimeDecoder.DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assert(t, inputs.A.Val.IsZero())

	e = withOptionalTimeDecoder.DecodeJSON(&inputs, []byte(`{"a":""}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assert(t, inputs.A.Val.IsZero())

	e = withOptionalTimeDecoder.DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assert(t, inputs.A.Val.IsZero())

	e = withOptionalTimeDecoder.DecodeMap(&inputs, map[string]interface{}{"a": time.Time{}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assert(t, inputs.A.Val.IsZero())
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

	e = NewDecoder(&inputs).DecodeMap(&inputs, map[string]interface{}{"a": time.Time{}})
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
}

type withOptionalNullTime struct {
	A Time `meta_null:"true"`
}

var withOptionalNullTimeDecoder = NewDecoder(&withOptionalNullTime{})

func TestOptionalNullTimeSuccess(t *testing.T) {
	var inputs withOptionalNullTime
	e := withOptionalNullTimeDecoder.DecodeValues(&inputs, url.Values{"a": {"2015-06-02T16:33:22Z"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, false)
	assert(t, inputs.A.Val.Equal(time.Date(2015, 6, 2, 16, 33, 22, 0, time.UTC)))

	inputs = withOptionalNullTime{}
	e = withOptionalNullTimeDecoder.DecodeJSON(&inputs, []byte(`{"a":"2015-06-02T16:33:22Z"}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, false)
	assert(t, inputs.A.Val.Equal(time.Date(2015, 6, 2, 16, 33, 22, 0, time.UTC)))
}

func TestOptionalNullTimeNull(t *testing.T) {
	var inputs withOptionalNullTime
	e := withOptionalNullTimeDecoder.DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, true)
	assert(t, inputs.A.Val.IsZero())

	inputs = withOptionalNullTime{}
	e = withOptionalNullTimeDecoder.DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, true)
	assert(t, inputs.A.Val.IsZero())

	e = withOptionalNullTimeDecoder.DecodeMap(&inputs, map[string]interface{}{"a": time.Time{}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, true)
	assert(t, inputs.A.Val.IsZero())
}

func TestOptionalNullTimeOmitted(t *testing.T) {
	var inputs withOptionalNullTime
	e := withOptionalNullTimeDecoder.DecodeValues(&inputs, url.Values{})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Null, false)
	assert(t, inputs.A.Val.IsZero())

	inputs = withOptionalNullTime{}
	e = withOptionalNullTimeDecoder.DecodeJSON(&inputs, []byte(`{}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Null, false)
	assert(t, inputs.A.Val.IsZero())
}

func assertTimeInRange(t *testing.T, value, start, end time.Time) {
	assert(t, value.Equal(start) || value.After(start))
	assert(t, value.Equal(end) || value.Before(end))
}

func TestTimeExpressions(t *testing.T) {
	for expression, assertion := range map[string]func(output, before, after time.Time){
		// Simple keywords
		"now": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before, after)
		},
		"today": func(output, before, after time.Time) {
			assertTimeInRange(t, output,
				time.Date(before.Year(), before.Month(), before.Day(), 0, 0, 0, 0, time.UTC),
				time.Date(after.Year(), after.Month(), after.Day(), 0, 0, 0, 0, time.UTC),
			)
		},
		"yesterday": func(output, before, after time.Time) {
			assertTimeInRange(t, output,
				time.Date(before.Year(), before.Month(), before.Day()-1, 0, 0, 0, 0, time.UTC),
				time.Date(after.Year(), after.Month(), after.Day()-1, 0, 0, 0, 0, time.UTC),
			)
		},
		"tomorrow": func(output, before, after time.Time) {
			assertTimeInRange(t, output,
				time.Date(before.Year(), before.Month(), before.Day()+1, 0, 0, 0, 0, time.UTC),
				time.Date(after.Year(), after.Month(), after.Day()+1, 0, 0, 0, 0, time.UTC),
			)
		},
		// Past expressions (<n>_<unit>_ago)
		"99_nanoseconds_ago": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.Add(-99*time.Nanosecond), after.Add(-99*time.Nanosecond))
		},
		"31_seconds_ago": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.Add(-31*time.Second), after.Add(-31*time.Second))
		},
		"1_minute_ago": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.Add(-time.Minute), after.Add(-time.Minute))
		},
		"48_hours_ago": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.Add(-48*time.Hour), after.Add(-48*time.Hour))
		},
		"1_day_ago": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.AddDate(0, 0, -1), after.AddDate(0, 0, -1))
		},
		"5_days_ago": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.AddDate(0, 0, -5), after.AddDate(0, 0, -5))
		},
		"3_weeks_ago": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.AddDate(0, 0, -21), after.AddDate(0, 0, -21))
		},
		"2_months_ago": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.AddDate(0, -2, 0), after.AddDate(0, -2, 0))
		},
		"4_years_ago": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.AddDate(-4, 0, 0), after.AddDate(-4, 0, 0))
		},
		// Future expressions (<n>_<unit>_from_now)
		"99_nanoseconds_from_now": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.Add(99*time.Nanosecond), after.Add(99*time.Nanosecond))
		},
		"31_seconds_from_now": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.Add(31*time.Second), after.Add(31*time.Second))
		},
		"1_minute_from_now": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.Add(time.Minute), after.Add(time.Minute))
		},
		"48_hours_from_now": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.Add(48*time.Hour), after.Add(48*time.Hour))
		},
		"1_day_from_now": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.AddDate(0, 0, 1), after.AddDate(0, 0, 1))
		},
		"5_days_from_now": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.AddDate(0, 0, 5), after.AddDate(0, 0, 5))
		},
		"3_weeks_from_now": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.AddDate(0, 0, 21), after.AddDate(0, 0, 21))
		},
		"2_months_from_now": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.AddDate(0, 2, 0), after.AddDate(0, 2, 0))
		},
		"4_years_from_now": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.AddDate(4, 0, 0), after.AddDate(4, 0, 0))
		},
	} {
		var inputs withTime

		before := time.Now()
		e := withTimeDecoder.DecodeValues(&inputs, url.Values{"a": {expression}})
		after := time.Now()
		assertEqual(t, e, ErrorHash(nil))
		assertEqual(t, inputs.A.Present, true)
		assertion(inputs.A.Val, before, after)
	}
}
