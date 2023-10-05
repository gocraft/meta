package meta

import (
	"reflect"
	"strings"
	"unicode/utf8"
)

type StringSlice struct {
	Val  []string
	Path string
}

type StringSliceOptions struct {
	*StringOptions
	*SliceOptions
}

func (i *StringSlice) ParseOptions(tag reflect.StructTag) interface{} {
	var tempS String
	opts := tempS.ParseOptions(tag)

	// unlike String we want Blank to default to true so we can clean up input values like: a,b,c,,,d
	stringOpts := opts.(*StringOptions)
	stringOpts.Blank = true
	if tag.Get("meta_blank") == "false" {
		stringOpts.Blank = false
	}

	return &StringSliceOptions{
		StringOptions: stringOpts,
		SliceOptions:  ParseSliceOptions(tag),
	}
}

func (n *StringSlice) JSONValue(path string, i interface{}, options interface{}) Errorable {
	n.Path = path
	n.Val = nil
	if i == nil {
		return ErrBlank
	}

	var errorsInSlice ErrorSlice
	switch value := i.(type) {
	case string:
		return n.FormValue(value, options)
	case []interface{}:
		if len(value) == 0 {
			return ErrBlank
		}
		opts := options.(*StringSliceOptions)
		stringOpts := opts.StringOptions
		sliceOpts := opts.SliceOptions

		if sliceOpts.MinLengthPresent && len(value) < sliceOpts.MinLength {
			return ErrMinLength
		}

		if sliceOpts.MaxLengthPresent && len(value) > sliceOpts.MaxLength {
			return ErrMaxLength
		}

		for _, v := range value {
			var s String
			if err := s.JSONValue("", v, stringOpts); err != nil {
				errorsInSlice = append(errorsInSlice, err)
				if err == ErrBlank && !opts.DiscardBlank {
					n.Val = append(n.Val, s.Val)
				}
			} else {
				if !opts.DiscardBlank || s.Val != "" {
					errorsInSlice = append(errorsInSlice, nil)
					n.Val = append(n.Val, s.Val)
				}
			}
		}
		if errorsInSlice.Len() > 0 {
			return errorsInSlice
		}
	}
	return nil

}

func (i *StringSlice) FormValue(value string, options interface{}) Errorable {
	if !utf8.ValidString(value) {
		return ErrUtf8
	}

	if value == "" {
		return ErrBlank
	}

	var tempS String

	opts := options.(*StringSliceOptions)
	stringOpts := opts.StringOptions
	sliceOpts := opts.SliceOptions

	strs := strings.Split(value, ",")

	if sliceOpts.MinLengthPresent && len(strs) < sliceOpts.MinLength {
		return ErrMinLength
	}

	if sliceOpts.MaxLengthPresent && len(strs) > sliceOpts.MaxLength {
		return ErrMaxLength
	}

	var errorsInSlice ErrorSlice

	for _, s := range strs {
		tempS.Val = ""

		if err := tempS.FormValue(s, stringOpts); err != nil {
			errorsInSlice = append(errorsInSlice, err)
			if err == ErrBlank && !opts.DiscardBlank {
				i.Val = append(i.Val, tempS.Val)
			}
		} else {
			if !opts.DiscardBlank || tempS.Val != "" {
				errorsInSlice = append(errorsInSlice, nil)
				i.Val = append(i.Val, tempS.Val)
			}
		}
	}

	if errorsInSlice.Len() > 0 {
		return errorsInSlice
	}

	return nil
}

func (s StringSlice) MarshalJSON() ([]byte, error) {
	if len(s.Val) > 0 {
		return MetaJson.Marshal(s.Val)
	}
	return nullString, nil
}
