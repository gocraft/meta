package meta

import (
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
	Presence
}

type StringOptions struct {
	Required        bool
	DiscardBlank    bool
	Strip           bool
	Blank           bool
	MinRunesPresent bool
	MinRunes        int
	MaxRunesPresent bool
	MaxRunes        int
	In              []string
}

func NewString(s string) String {
	return String{s, Presence{Present: true}}
}

func (s *String) ParseOptions(tag reflect.StructTag) interface{} {
	opts := &StringOptions{
		Required:        false,
		DiscardBlank:    true,
		Strip:           true,
		Blank:           false,
		MinRunesPresent: false,
		MinRunes:        0,
		MaxRunesPresent: false,
		MaxRunes:        0,
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

	if in := tag.Get("meta_in"); in != "" {
		for _, s := range strings.Split(in, ",") {
			opts.In = append(opts.In, strings.TrimSpace(s))
		}
	}

	return opts
}

func (s *String) JSONValue(i interface{}, options interface{}) Errorable {
	if i == nil {
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
	if s.Present {
		return s.Val, nil
	}
	return nil, nil
}

func (s String) MarshalJSON() ([]byte, error) {
	if s.Present {
		return json.Marshal(s.Val)
	}
	return nullString, nil
}
