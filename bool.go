package meta

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"reflect"
	"strconv"
)

//
// Bool
//

type Bool struct {
	Val bool
	Nullity
	Presence
	Path string
}

type BoolOptions struct {
	Required     bool
	DiscardBlank bool
	Null         bool
}

func NewBool(b bool) Bool {
	return Bool{b, Nullity{false}, Presence{true}, ""}
}

func (b *Bool) ParseOptions(tag reflect.StructTag) interface{} {
	opts := &BoolOptions{
		Required:     false,
		DiscardBlank: true,
		Null:         false,
	}

	if tag.Get("meta_required") == "true" {
		opts.Required = true
	}

	if tag.Get("meta_discard_blank") == "false" {
		opts.DiscardBlank = false
	}

	if tag.Get("meta_null") == "true" {
		opts.Null = true
	}

	return opts
}

func (b *Bool) FormValue(value string, options interface{}) Errorable {
	opts := options.(*BoolOptions)

	if value == "" {
		if opts.Null {
			b.Present = true
			b.Null = true
			return nil
		}
		if opts.Required {
			return ErrBlank
		}
		if !opts.DiscardBlank {
			b.Present = true
			return ErrBlank
		}
		return nil
	}

	if v, err := strconv.ParseBool(value); err == nil {
		b.Val = v
		b.Present = true
		return nil
	}

	return ErrBool
}

func (b *Bool) JSONValue(path string, i interface{}, options interface{}) Errorable {
	opts := options.(*BoolOptions)
	b.Path = path

	if i == nil {
		if opts.Null {
			b.Present = true
			b.Null = true
			return nil
		}
		if opts.Required || !opts.DiscardBlank {
			return ErrBlank
		}
		return nil
	}

	switch value := i.(type) {
	case string:
		return b.FormValue(value, options)
	case json.Number:
		return b.FormValue(string(value), options)
	case bool:
		b.Val = value
		b.Present = true
		return nil
	}

	return ErrBool
}

func (b Bool) Value() (driver.Value, error) {
	if b.Present && !b.Null {
		return b.Val, nil
	}
	return nil, nil
}

func (b Bool) MarshalJSON() ([]byte, error) {
	if b.Present && !b.Null {
		return MetaJson.Marshal(b.Val)
	}
	return nullString, nil
}

func (b *Bool) UnmarshalJSON(bs []byte) error {
	if bytes.Equal(nullString, bs) {
		b.Nullity = Nullity{true}
		return nil
	}

	err := MetaJson.Unmarshal(bs, &b.Val)
	if err != nil {
		return err
	}
	b.Presence = Presence{true}
	return nil
}
