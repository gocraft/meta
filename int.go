package meta

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
)

//
// Int64, Uint64
//

type Int64 struct {
	Val int64
	Nullity
	Presence
	Path string
}

type Uint64 struct {
	Val uint64
	Nullity
	Presence
	Path string
}

type IntOptions struct {
	Required     bool
	Null         bool
	DiscardBlank bool
	MinPresent   bool
	Min          int64
	MaxPresent   bool
	Max          int64
	In           []int64
}

type UintOptions struct {
	Required     bool
	Null         bool
	DiscardBlank bool
	MinPresent   bool
	Min          uint64
	MaxPresent   bool
	Max          uint64
	In           []uint64
}

func NewInt64(val int64) Int64 {
	return Int64{val, Nullity{false}, Presence{true}, ""}
}

func NewUint64(val uint64) Uint64 {
	return Uint64{val, Nullity{false}, Presence{true}, ""}
}

func (i *Int64) ParseOptions(tag reflect.StructTag) interface{} {
	opts := &IntOptions{
		DiscardBlank: true,
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

	if nstr := tag.Get("meta_min"); nstr != "" {
		n, err := strconv.ParseInt(nstr, 10, 64)
		if err != nil {
			panic(err.Error())
		}

		opts.MinPresent = true
		opts.Min = n
	}

	if nstr := tag.Get("meta_max"); nstr != "" {
		n, err := strconv.ParseInt(nstr, 10, 64)
		if err != nil {
			panic(err.Error())
		}

		opts.MaxPresent = true
		opts.Max = n
	}

	if in := tag.Get("meta_in"); in != "" {
		for _, s := range strings.Split(in, ",") {
			n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
			if err != nil {
				panic(err.Error())
			}

			opts.In = append(opts.In, n)
		}
	}

	return opts
}

func (i *Uint64) ParseOptions(tag reflect.StructTag) interface{} {
	opts := &UintOptions{
		DiscardBlank: true,
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

	if nstr := tag.Get("meta_min"); nstr != "" {
		n, err := strconv.ParseUint(nstr, 10, 64)
		if err != nil {
			panic(err.Error())
		}

		opts.MinPresent = true
		opts.Min = n
	}

	if nstr := tag.Get("meta_max"); nstr != "" {
		n, err := strconv.ParseUint(nstr, 10, 64)
		if err != nil {
			panic(err.Error())
		}

		opts.MaxPresent = true
		opts.Max = n
	}

	if in := tag.Get("meta_in"); in != "" {
		for _, s := range strings.Split(in, ",") {
			n, err := strconv.ParseUint(strings.TrimSpace(s), 10, 64)
			if err != nil {
				panic(err.Error())
			}

			opts.In = append(opts.In, n)
		}
	}

	return opts
}

func (n *Int64) JSONValue(path string, i interface{}, options interface{}) Errorable {
	n.Path = path
	if i == nil {
		return n.FormValue("", options)
	}

	switch value := i.(type) {
	case int:
		return n.validateValue(int64(value), options)
	case int64:
		return n.validateValue(value, options)
	case json.Number:
		return n.FormValue(string(value), options)
	case string:
		return n.FormValue(value, options)
	}

	return ErrInt
}

func (i *Int64) FormValue(value string, options interface{}) Errorable {
	opts := options.(*IntOptions)

	if value == "" {
		if opts.Null {
			i.Null = true
			i.Present = true
			return nil
		}
		if opts.Required {
			return ErrBlank
		}
		if !opts.DiscardBlank {
			i.Present = true
			return ErrBlank
		}
		return nil
	}

	if n, err := strconv.ParseInt(value, 10, 64); err == nil {
		return i.validateValue(n, options)
	} else {
		numError := err.(*strconv.NumError)
		if numError.Err == strconv.ErrRange {
			return ErrIntRange
		} else {
			return ErrInt
		}
	}
	return nil
}

func (i *Int64) validateValue(value int64, options interface{}) Errorable {
	opts := options.(*IntOptions)

	if opts.MinPresent && value < opts.Min {
		return ErrMin
	}
	if opts.MaxPresent && value > opts.Max {
		return ErrMax
	}
	if len(opts.In) > 0 {
		found := false
		for _, i := range opts.In {
			if i == value {
				found = true
			}
		}
		if !found {
			return ErrIn
		}
	}

	i.Val = value
	i.Present = true
	return nil
}

func (n *Uint64) JSONValue(path string, i interface{}, options interface{}) Errorable {
	n.Path = path
	if i == nil {
		return n.FormValue("", options)
	}

	switch value := i.(type) {
	case int:
		return n.validateValue(uint64(value), options)
	case int64:
		return n.validateValue(uint64(value), options)
	case uint64:
		return n.validateValue(value, options)
	case json.Number:
		return n.FormValue(string(value), options)
	case string:
		return n.FormValue(value, options)
	}

	return ErrInt
}

func (i *Uint64) FormValue(value string, options interface{}) Errorable {
	opts := options.(*UintOptions)

	if value == "" {
		if opts.Null {
			i.Null = true
			i.Present = true
			return nil
		}
		if opts.Required {
			return ErrBlank
		}
		if !opts.DiscardBlank {
			i.Present = true
			return ErrBlank
		}
		return nil
	}

	if n, err := strconv.ParseUint(value, 10, 64); err == nil {
		return i.validateValue(n, options)
	} else {
		numError := err.(*strconv.NumError)

		if numError.Err == strconv.ErrRange {
			return ErrIntRange
		} else {
			return ErrInt
		}
	}
	return nil
}

func (i *Uint64) validateValue(value uint64, options interface{}) Errorable {
	opts := options.(*UintOptions)

	if opts.MinPresent && value < opts.Min {
		return ErrMin
	}
	if opts.MaxPresent && value > opts.Max {
		return ErrMax
	}
	if len(opts.In) > 0 {
		found := false
		for _, i := range opts.In {
			if i == value {
				found = true
			}
		}
		if !found {
			return ErrIn
		}
	}

	i.Val = value
	i.Present = true
	return nil
}

func (i Int64) Value() (driver.Value, error) {
	if i.Present && !i.Null {
		return int64(i.Val), nil
	}
	return nil, nil
}

// NOTE: I know I am casting uints to int64's. This is per Go's docs, which does NOT list uint64 as a viable type. Unsure what that means for a large Uint64.
func (i Uint64) Value() (driver.Value, error) {
	if i.Present && !i.Null {
		return int64(i.Val), nil
	}
	return nil, nil
}

func (i Int64) MarshalJSON() ([]byte, error) {
	if i.Present && !i.Null {
		return MetaJson.Marshal(i.Val)
	}
	return nullString, nil
}

func (i *Int64) UnmarshalJSON(b []byte) error {
	if bytes.Equal(nullString, b) {
		i.Nullity = Nullity{true}
		return nil
	}
	err := MetaJson.Unmarshal(b, &i.Val)
	if err != nil {
		return err
	}
	i.Presence = Presence{true}
	i.Nullity = Nullity{false}
	return nil
}

func (i Uint64) MarshalJSON() ([]byte, error) {
	if i.Present && !i.Null {
		return MetaJson.Marshal(i.Val)
	}
	return nullString, nil
}

func (i *Uint64) UnmarshalJSON(b []byte) error {
	if bytes.Equal(nullString, b) {
		i.Nullity = Nullity{true}
		return nil
	}

	err := MetaJson.Unmarshal(b, &i.Val)
	if err != nil {
		return err
	}
	i.Presence = Presence{true}
	i.Nullity = Nullity{false}
	return nil
}
