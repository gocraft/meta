package meta

import (
	"database/sql/driver"
	"encoding/json"
	"reflect"
	"time"
)

//
// Time
//

type Time struct {
	Val time.Time
	Nullity
	Presence
}

type TimeOptions struct {
	Required     bool
	DiscardBlank bool
	Null         bool
	Format       []string
}

func NewTime(t time.Time) Time {
	return Time{t, Nullity{false}, Presence{true}}
}

func (t *Time) ParseOptions(tag reflect.StructTag) interface{} {
	opts := &TimeOptions{
		Required:     false,
		DiscardBlank: true,
		Null:         false,
		Format:       []string{time.RFC3339},
	}

	if tag.Get("meta_required") == "true" {
		opts.Required = true
	}

	if tag.Get("meta_null") == "true" {
		opts.Null = true
	}

	if tag.Get("meta_discard_blank") == "false" {
		opts.DiscardBlank = false
	}

	if tag.Get("meta_format") != "" {
		opts.Format = []string{tag.Get("meta_format")}
	}

	return opts
}

func (t *Time) JSONValue(i interface{}, options interface{}) Errorable {
	if i == nil {
		return t.FormValue("", options)
	}

	switch value := i.(type) {
	case string:
		return t.FormValue(value, options)
	}

	return ErrTime
}

func (t *Time) FormValue(value string, options interface{}) Errorable {
	opts := options.(*TimeOptions)

	if value == "" {
		if opts.Null {
			t.Present = true
			t.Null = true
			return nil
		}
		if opts.Required {
			return ErrBlank
		}
		if !opts.DiscardBlank {
			t.Present = true
			return ErrBlank
		}
		return nil
	}

	for _, format := range opts.Format {
		if v, err := time.Parse(format, value); err == nil {
			t.Val = v
			t.Present = true
			return nil
		}
	}

	return ErrTime
}

func (t Time) Value() (driver.Value, error) {
	if t.Present && !t.Null {
		return t.Val, nil
	}
	return nil, nil
}

func (t Time) MarshalJSON() ([]byte, error) {
	if t.Present && !t.Null {
		return json.Marshal(t.Val)
	}
	return nullString, nil
}
