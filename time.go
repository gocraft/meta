package meta

import (
	"bytes"
	"database/sql/driver"
	"reflect"
	"regexp"
	"strconv"
	"time"
)

//
// Time
//

type Time struct {
	Val time.Time
	Nullity
	Presence
	Path string
}

type TimeOptions struct {
	Required     bool
	DiscardBlank bool
	Null         bool
	Format       []string
}

func NewTime(t time.Time) Time {
	return Time{t, Nullity{false}, Presence{true}, ""}
}

func (t *Time) ParseOptions(tag reflect.StructTag) interface{} {
	opts := &TimeOptions{
		Required:     false,
		DiscardBlank: true,
		Null:         false,
		Format:       []string{time.RFC3339, "expression"},
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

func (t *Time) JSONValue(path string, i interface{}, options interface{}) Errorable {
	t.Path = path
	if i == nil {
		return t.FormValue("", options)
	}

	switch value := i.(type) {
	case time.Time:
		if value.IsZero() {
			opts := options.(*TimeOptions)
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
		t.Present = true
		t.Val = value
		return nil
	case string:
		return t.FormValue(value, options)
	}

	return ErrTime
}

type expressionParser struct {
	*regexp.Regexp
	Parse func([]string) (time.Time, bool)
}

var timeExpressionParsers = []expressionParser{
	{
		Regexp: regexp.MustCompile(`^(\d+)_(year|month|week|day|hour|minute|second|nanosecond)s?_(ago|from_now)$`),
		Parse: func(matches []string) (time.Time, bool) {
			delta, err := strconv.Atoi(matches[1])
			if err != nil {
				return time.Time{}, false
			}
			if matches[3] == "ago" {
				delta = -delta
			}
			switch matches[2] {
			case "year":
				return time.Now().AddDate(delta, 0, 0), true
			case "month":
				return time.Now().AddDate(0, delta, 0), true
			case "week":
				return time.Now().AddDate(0, 0, delta*7), true
			case "day":
				return time.Now().AddDate(0, 0, delta), true
			case "hour":
				return time.Now().Add(time.Duration(delta) * time.Hour), true
			case "minute":
				return time.Now().Add(time.Duration(delta) * time.Minute), true
			case "second":
				return time.Now().Add(time.Duration(delta) * time.Second), true
			case "nanosecond":
				return time.Now().Add(time.Duration(delta) * time.Nanosecond), true
			}

			return time.Time{}, false
		},
	},
	{
		Regexp: regexp.MustCompile(`^now$`),
		Parse: func(matches []string) (time.Time, bool) {
			return time.Now(), true
		},
	},
	{
		Regexp: regexp.MustCompile(`^today$`),
		Parse: func(matches []string) (time.Time, bool) {
			return time.Now().Truncate(24 * time.Hour), true
		},
	},
	{
		Regexp: regexp.MustCompile(`^yesterday$`),
		Parse: func(matches []string) (time.Time, bool) {
			return time.Now().Truncate(24*time.Hour).AddDate(0, 0, -1), true
		},
	},
	{
		Regexp: regexp.MustCompile(`^tomorrow$`),
		Parse: func(matches []string) (time.Time, bool) {
			return time.Now().Truncate(24*time.Hour).AddDate(0, 0, 1), true
		},
	},
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
		switch format {
		case "expression":
			for _, parser := range timeExpressionParsers {
				submatches := parser.Regexp.FindStringSubmatch(value)
				if len(submatches) == 0 {
					continue
				}
				if v, ok := parser.Parse(submatches); ok {
					t.Val = v
					t.Present = true
					return nil
				}
			}
		default:
			if v, err := time.Parse(format, value); err == nil {
				t.Val = v
				t.Present = true
				return nil
			}
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
		return MetaJson.Marshal(t.Val)
	}
	return nullString, nil
}

func (t *Time) UnmarshalJSON(b []byte) error {
	if bytes.Equal(nullString, b) {
		t.Nullity = Nullity{true}
		return nil
	}
	err := MetaJson.Unmarshal(b, &t.Val)
	if err != nil {
		return err
	}
	t.Presence = Presence{true}
	t.Nullity = Nullity{false}
	return nil
}
