package meta

import (
	"reflect"
	"strings"
)

//
// Int64Slice, TODO: Uint64Slice
//

type Int64Slice struct {
	Val  []int64
	Path string
}

type IntSliceOptions struct {
	*IntOptions
	*SliceOptions
}

func (i *Int64Slice) ParseOptions(tag reflect.StructTag) interface{} {
	var tempI Int64
	opts := tempI.ParseOptions(tag)
	return &IntSliceOptions{
		IntOptions:   opts.(*IntOptions),
		SliceOptions: ParseSliceOptions(tag),
	}
}

func (n *Int64Slice) JSONValue(path string, i interface{}, options interface{}) Errorable {
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
		opts := options.(*IntSliceOptions)
		intOpts := opts.IntOptions
		sliceOpts := opts.SliceOptions

		if sliceOpts.MinLengthPresent && len(value) < sliceOpts.MinLength {
			return ErrMinLength
		}

		if sliceOpts.MaxLengthPresent && len(value) > sliceOpts.MaxLength {
			return ErrMaxLength
		}

		for _, v := range value {
			var num Int64
			if err := num.JSONValue("", v, intOpts); err != nil {
				errorsInSlice = append(errorsInSlice, err)
			} else {
				errorsInSlice = append(errorsInSlice, nil)
				n.Val = append(n.Val, num.Val)
			}
		}
		if errorsInSlice.Len() > 0 {
			return errorsInSlice
		}
	}
	return nil
}

func (i *Int64Slice) FormValue(value string, options interface{}) Errorable {
	if value == "" {
		return ErrBlank
	}

	var tempI Int64

	opts := options.(*IntSliceOptions)
	intOpts := opts.IntOptions
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
		tempI.Val = 0
		if err := tempI.FormValue(s, intOpts); err != nil {
			errorsInSlice = append(errorsInSlice, err)
		} else {
			errorsInSlice = append(errorsInSlice, nil)
			i.Val = append(i.Val, tempI.Val)
		}
	}

	if errorsInSlice.Len() > 0 {
		return errorsInSlice
	}

	return nil
}

func (s Int64Slice) MarshalJSON() ([]byte, error) {
	if len(s.Val) > 0 {
		return MetaJson.Marshal(s.Val)
	}
	return nullString, nil
}
