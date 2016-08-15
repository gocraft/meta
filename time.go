package meta

import (
	"database/sql/driver"
	"reflect"
	"time"
)

//
// Time
//

type Time struct {
	Val time.Time
	Presence
}

type TimeOptions struct {
	Required     bool
	DiscardBlank bool
	Format       string
}

func NewTime(t time.Time) Time {
	return Time{t, Presence{true}}
}

func (t *Time) ParseOptions(tag reflect.StructTag) interface{} {
	opts := &TimeOptions{
		Required:     false,
		DiscardBlank: true,
		Format:       time.RFC3339,
	}

	if tag.Get("meta_required") == "true" {
		opts.Required = true
	}

	if tag.Get("meta_discard_blank") == "false" {
		opts.DiscardBlank = false
	}

	if tag.Get("meta_format") != "" {
		opts.Format = tag.Get("meta_format")
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
		if opts.Required {
			return ErrBlank
		}
		if !opts.DiscardBlank {
			t.Present = true
			return ErrBlank
		}
		return nil
	}

	if v, err := time.Parse(opts.Format, value); err == nil {
		t.Val = v
		t.Present = true
		return nil
	}

	return ErrTime
}

func (t Time) Value() (driver.Value, error) {
	if t.Present {
		return t.Val, nil
	}
	return nil, nil
}
