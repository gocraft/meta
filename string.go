package meta

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

//
// String
//

type String struct {
	Val string
	Nullity
	Presence
	Path string
}

type StringOptions struct {
	Required        bool
	DiscardBlank    bool
	Strip           bool
	Blank           bool
	Null            bool
	MinRunesPresent bool
	MinRunes        int
	MaxRunesPresent bool
	MaxRunes        int
	MaxBytesPresent bool
	MaxBytes        int
	In              []string
}

func NewString(s string) String {
	return String{s, Nullity{false}, Presence{true}, ""}
}

func (s *String) ParseOptions(tag reflect.StructTag) interface{} {
	opts := &StringOptions{
		Required:        false,
		DiscardBlank:    true,
		Strip:           true,
		Blank:           false,
		Null:            false,
		MinRunesPresent: false,
		MinRunes:        0,
		MaxRunesPresent: false,
		MaxRunes:        0,
		MaxBytesPresent: false,
		MaxBytes:        0,
	}

	// need this here to implement discard_blank
	if tag.Get("meta_required") == "true" {
		opts.Required = true
	}

	if tag.Get("meta_discard_blank") == "false" {
		opts.DiscardBlank = false
	}

	if tag.Get("meta_strip") == "false" {
		opts.Strip = false
	}

	if tag.Get("meta_blank") == "true" {
		opts.Blank = true
	}

	if tag.Get("meta_null") == "true" {
		opts.Null = true
	}

	if minRunesString := tag.Get("meta_min_runes"); minRunesString != "" {
		minRunes, err := strconv.ParseInt(minRunesString, 10, 0)
		if err != nil {
			panic(err.Error())
		}

		opts.MinRunesPresent = true
		opts.MinRunes = int(minRunes)
	}

	if maxRunesString := tag.Get("meta_max_runes"); maxRunesString != "" {
		maxRunes, err := strconv.ParseInt(maxRunesString, 10, 0)
		if err != nil {
			panic(err.Error())
		}

		opts.MaxRunesPresent = true
		opts.MaxRunes = int(maxRunes)
	}

	if maxBytesString := tag.Get("meta_max_bytes"); maxBytesString != "" {
		maxBytes, err := strconv.ParseInt(maxBytesString, 10, 0)
		if err != nil {
			panic(err.Error())
		}

		opts.MaxBytesPresent = true
		opts.MaxBytes = int(maxBytes)
	}

	if in := tag.Get("meta_in"); in != "" {
		for _, s := range strings.Split(in, ",") {
			opts.In = append(opts.In, strings.TrimSpace(s))
		}
	}

	return opts
}

func (s *String) JSONValue(path string, i interface{}, options interface{}) Errorable {
	s.Path = path
	if i == nil {
		opts := options.(*StringOptions)
		if opts.Null {
			s.Present = true
			s.Null = true
			return nil
		}
		return s.FormValue("", options)
	}

	switch value := i.(type) {
	case string:
		return s.FormValue(value, options)
	case bool:
		return s.FormValue(fmt.Sprint(i), options)
	case json.Number:
		return s.FormValue(fmt.Sprint(i), options)
	}
	return ErrString
}

func (s *String) FormValue(value string, options interface{}) Errorable {
	if !utf8.ValidString(value) {
		return ErrUtf8
	}

	opts := options.(*StringOptions)

	// strip
	if opts.Strip {
		value = strings.TrimSpace(value)
	}

	runeCount := utf8.RuneCountInString(value)

	if runeCount == 0 {
		if opts.Blank {
			s.Present = true
			return nil
		}
		if opts.Null {
			s.Present = true
			s.Null = true
			return nil
		}
		if opts.Required {
			return ErrBlank
		}
		if !opts.DiscardBlank {
			s.Present = true
			return ErrBlank
		}
		return nil
	}

	// min_runes
	if opts.MinRunesPresent {
		if runeCount < opts.MinRunes {
			return ErrMinRunes
		}
	}

	// max_runes
	if opts.MaxRunesPresent {
		if runeCount > opts.MaxRunes {
			return ErrMaxRunes
		}
	}

	if opts.MaxBytesPresent {
		if len(value) > opts.MaxBytes {
			return ErrMaxBytes
		}
	}

	// in
	if len(opts.In) > 0 {
		found := false
		for _, v := range opts.In {
			if v == value {
				found = true
			}
		}
		if !found {
			return ErrIn
		}
	}

	// success
	s.Val = value
	s.Present = true

	return nil
}

func (s String) Value() (driver.Value, error) {
	if s.Present && !s.Null {
		return s.Val, nil
	}
	return nil, nil
}

func (s String) MarshalJSON() ([]byte, error) {
	if s.Present && !s.Null {
		return MetaJson.Marshal(s.Val)
	}
	return nullString, nil
}

func (s *String) UnmarshalJSON(bs []byte) error {
	if bytes.Equal(nullString, bs) {
		s.Nullity = Nullity{true}
		return nil
	}

	err := MetaJson.Unmarshal(bs, &s.Val)
	if err != nil {
		return err
	}

	s.Nullity = Nullity{false}
	s.Presence = Presence{true}
	return nil
}
