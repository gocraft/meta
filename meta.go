package meta

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"reflect"
	"strconv"
)

type Valuer interface {
	ParseOptions(tag reflect.StructTag) interface{}
	JSONValue(path string, value interface{}, options interface{}) Errorable
}

var (
	reflectTypeValuer = reflect.TypeOf((*Valuer)(nil)).Elem()
)

type Optionaler interface {
	Optional() bool
}

type decoderFieldCategory int

const (
	categoryValuer decoderFieldCategory = iota
	categoryStruct
	categorySliceOfValues
	categorySliceOfStructs
	categoryAllFieldsMap
)

var nullString = []byte("null")

type SliceOptions struct {
	MinLengthPresent bool
	MinLength        int
	MaxLengthPresent bool
	MaxLength        int
}

func ParseSliceOptions(tag reflect.StructTag) *SliceOptions {
	sliceOpts := &SliceOptions{}

	if minLengthString := tag.Get("meta_min_length"); minLengthString != "" {
		minLength, err := strconv.ParseInt(minLengthString, 10, 0)
		if err != nil {
			panic(err.Error())
		}

		sliceOpts.MinLengthPresent = true
		sliceOpts.MinLength = int(minLength)
	}

	if maxLengthString := tag.Get("meta_max_length"); maxLengthString != "" {
		maxLength, err := strconv.ParseInt(maxLengthString, 10, 0)
		if err != nil {
			panic(err.Error())
		}

		sliceOpts.MaxLengthPresent = true
		sliceOpts.MaxLength = int(maxLength)
	}

	return sliceOpts
}

type DecoderField struct {
	Name            string // key in the input
	Required        bool
	DiscardInvalid  bool
	Options         interface{}
	needsAllocation bool // true if we need to reflect.New
	Default         string
	Doc             string
	DocPattern      string

	*SliceOptions

	// The type of field it is:
	fieldCategory decoderFieldCategory
	StructDecoder *Decoder // If the field is a nested struct or a slice of nested structs, this is set to the decoder.

	fieldIndex []int // Given the struct Value, how can we get the field with .FieldByIndex(fieldIndex)

	// Basic type information:
	fieldType      reflect.Type // Type of the field. Eg, TypeOf(field)
	fieldKind      reflect.Kind
	indirectedType reflect.Type
	indirectedKind reflect.Kind

	// ElemXxx: Applies to Slices.
	// elemType is the TypeOf each slice element. If that's a pointer, then Indirected
	// It can be the case that elemType == elemIndirectedType.
	elemType           reflect.Type
	elemKind           reflect.Kind
	elemIndirectedType reflect.Type
	elemIndirectedKind reflect.Kind
}

type Decoder struct {
	StructType reflect.Type
	Fields     []DecoderField
	Options    DecoderOptions
}

type DecoderOptions struct {
	TimeFormats []string
}

func NewDecoderWithOptions(destStruct interface{}, options DecoderOptions) *Decoder {
	destValue := reflect.ValueOf(destStruct)
	indirectedDest := reflect.Indirect(destValue)
	destType := indirectedDest.Type()

	if destValue.Kind() == reflect.Ptr && indirectedDest.Kind() == reflect.Struct {
		// we're good
	} else if destValue.Kind() == reflect.Struct {
		destType = destValue.Type()
		destValue = reflect.New(destType)
		indirectedDest = reflect.Indirect(destValue)
	} else {
		panic(fmt.Sprintf("expect ptr to struct or struct, got %s", destValue.Kind()))
	}

	decoder := &Decoder{StructType: destType}

	fieldCount := indirectedDest.NumField()
	for i := 0; i < fieldCount; i += 1 {
		field := indirectedDest.Field(i)
		fieldStruct := destType.Field(i) // type: StructField
		fieldType := field.Type()
		fieldKind := fieldType.Kind()

		var indirectedType reflect.Type
		var indirectedKind reflect.Kind
		if fieldKind == reflect.Ptr {
			indirectedType = fieldType.Elem()
			indirectedKind = indirectedType.Kind()
		} else {
			indirectedType = fieldType
			indirectedKind = fieldKind
		}

		var fieldInterface interface{} // This is going to be a pointer to a struct
		var needsAllocation bool
		if fieldKind == reflect.Struct {
			fieldInterface = field.Addr().Interface()
		} else if fieldKind == reflect.Ptr && indirectedKind == reflect.Struct {
			fieldInterface = reflect.New(indirectedType).Interface()
			needsAllocation = true
		}

		// Determine the key we're expecting in input
		metaName := fieldStruct.Tag.Get("meta")
		if metaName == "-" {
			continue
		} else if metaName == "" {
			metaName = NameMapping(fieldStruct.Name)
		}

		// Determine if it's required..
		required := fieldStruct.Tag.Get("meta_required") == "true"

		if fieldStruct.Anonymous && indirectedKind == reflect.Struct {
			// It's an embedded struct:
			embeddedDecoder := NewDecoderWithOptions(fieldInterface, options)

			for _, embeddedDField := range embeddedDecoder.Fields {
				idx := []int{i}
				idx = append(idx, embeddedDField.fieldIndex...)
				embeddedDField.fieldIndex = idx
				decoder.Fields = append(decoder.Fields, embeddedDField)
			}
		} else {
			dfield := DecoderField{
				Name:            metaName,
				Required:        required,
				needsAllocation: needsAllocation,
				fieldIndex:      []int{i},
				fieldType:       fieldType,
				fieldKind:       fieldKind,
				indirectedType:  indirectedType,
				indirectedKind:  indirectedKind,
			}

			dfield.Doc = fieldStruct.Tag.Get("doc")
			dfield.DocPattern = fieldStruct.Tag.Get("doc_pattern")

			// Determine what kind of field it is.
			if metaName == "*" && indirectedKind == reflect.Map {
				dfield.fieldCategory = categoryAllFieldsMap
			} else if valuer, ok := fieldInterface.(Valuer); ok {
				dfield.fieldCategory = categoryValuer
				dfield.Options = getParsedOptions(valuer, fieldStruct, options)
				if def := fieldStruct.Tag.Get("meta_default"); def != "" {
					dfield.Default = def
				}
				dfield.DiscardInvalid = fieldStruct.Tag.Get("meta_discard_invalid") == "true"
			} else if indirectedKind == reflect.Struct {
				dfield.fieldCategory = categoryStruct
				dfield.StructDecoder = NewDecoderWithOptions(fieldInterface, options)
			} else if indirectedKind == reflect.Slice {
				var elemType, elemIndirectedType reflect.Type
				var elemKind, elemIndirectedKind reflect.Kind
				elemType = fieldType.Elem()
				elemKind = elemType.Kind()

				if elemKind == reflect.Ptr {
					elemIndirectedType = elemType.Elem()
					elemIndirectedKind = elemIndirectedType.Kind()
				} else {
					elemIndirectedType = elemType
					elemIndirectedKind = elemKind
				}

				dfield.elemType = elemType
				dfield.elemKind = elemKind
				dfield.elemIndirectedType = elemIndirectedType
				dfield.elemIndirectedKind = elemIndirectedKind

				// Set slice validation options
				dfield.SliceOptions = ParseSliceOptions(fieldStruct.Tag)

				if reflect.PtrTo(elemIndirectedType).Implements(reflectTypeValuer) {
					dfield.fieldCategory = categorySliceOfValues
					valuer := reflect.New(elemIndirectedType).Interface().(Valuer) // Make a new object so we can use it to parse values.
					dfield.Options = getParsedOptions(valuer, fieldStruct, options)
				} else if elemIndirectedKind == reflect.Struct {
					dfield.fieldCategory = categorySliceOfStructs
					if elemIndirectedType == destType {
						dfield.StructDecoder = decoder
					} else {
						dfield.StructDecoder = NewDecoderWithOptions(reflect.New(elemIndirectedType).Interface(), options)
					}
				} else {
					panic("unknown type of slice")
				}
			}

			decoder.Fields = append(decoder.Fields, dfield)
		}
	}

	return decoder
}

func getParsedOptions(valuer Valuer, fieldStruct reflect.StructField, options DecoderOptions) interface{} {
	parsedOptions := valuer.ParseOptions(fieldStruct.Tag)
	if timeOptions, ok := parsedOptions.(*TimeOptions); ok && len(options.TimeFormats) > 0 {
		timeOptions.Format = options.TimeFormats
		parsedOptions = timeOptions
	}

	return parsedOptions
}

func NewDecoder(destStruct interface{}) *Decoder {
	return NewDecoderWithOptions(destStruct, DecoderOptions{})
}

func (d *Decoder) Decode(dest interface{}, values url.Values, b []byte) ErrorHash {
	return d.decode(reflect.ValueOf(dest), newMergedSource(newJSONSource(b), newFormValueSource(values)))
}

func (d *Decoder) DecodeJSON(dest interface{}, b []byte) ErrorHash {
	return d.Decode(dest, nil, b)
}

func (d *Decoder) DecodeValues(dest interface{}, values url.Values) ErrorHash {
	return d.Decode(dest, values, nil)
}

func (d *Decoder) DecodeMap(dest interface{}, m map[string]interface{}) ErrorHash {
	return d.decode(reflect.ValueOf(dest), newMapSource(m))
}

func (d *Decoder) decode(destValue reflect.Value, src source) ErrorHash {
	var errs ErrorHash

	indirectedDest := reflect.Indirect(destValue) // This should be the value of the struct

	if destValue.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("expect ptr, got %s", destValue.Kind()))
	}

	if indirectedDest.Type() != d.StructType {
		panic(fmt.Sprintf("expect type %s, got %s", d.StructType, indirectedDest.Type()))
	}

	for _, dfield := range d.Fields {
		fieldValue := indirectedDest.FieldByIndex(dfield.fieldIndex)

		metaName := dfield.Name

		switch dfield.fieldCategory {
		case categoryValuer:
			nestedValues := src.Get(metaName)
			if nestedValues.Malformed() {
				return ErrorHash{
					"error": ErrMalformed,
				}
			}

			ok := !nestedValues.Empty()
			var val interface{}
			if ok {
				nestedValues.Value(&val)
			} else if dfield.Default != "" {
				val = dfield.Default
				ok = true
			}
			if ok {
				valuerValue := fieldValue.Addr()
				var err Errorable
				if dfield.needsAllocation {
					fieldValue.Set(reflect.New(dfield.indirectedType))
					valuerValue = fieldValue
				}
				err = valuerValue.Interface().(Valuer).JSONValue(nestedValues.Path(), val, dfield.Options)
				if err != nil && !dfield.DiscardInvalid {
					errs = addError(errs, metaName, err)
				}
			} else if dfield.Required {
				errs = addError(errs, metaName, ErrRequired)
			}
		case categoryStruct:
			// Construct nestedValues
			// if the struct name is like FooBar,
			// {foo_bar.x=1, foo_bar.y=2} -> {x=1, y=2}
			nestedValues := src.Get(metaName)
			if nestedValues.Malformed() {
				return ErrorHash{
					"error": ErrMalformed,
				}
			}

			if !nestedValues.Empty() {
				var err ErrorHash
				if dfield.needsAllocation {
					fieldValue.Set(reflect.New(dfield.indirectedType))
					err = dfield.StructDecoder.decode(fieldValue, nestedValues)
				} else {
					err = dfield.StructDecoder.decode(fieldValue.Addr(), nestedValues)
				}
				if err != nil {
					errs = addError(errs, metaName, err)
				}
			} else if dfield.Required {
				errs = addError(errs, metaName, ErrRequired)
			}
		case categorySliceOfValues:
			sliceValue := fieldValue
			var errorsInSlice ErrorSlice

			sliceSrc := src.Get(metaName)
			for i := 0; true; i += 1 {
				nestedValues := sliceSrc.Get(fmt.Sprint(i)) // foo_bar.0, foo_bar.1, ...
				if nestedValues.Malformed() {
					return ErrorHash{
						"error": ErrMalformed,
					}
				}
				if nestedValues.Empty() {
					break
				}
				var val interface{}
				nestedValues.Value(&val)
				elPtrValue := reflect.New(dfield.elemIndirectedType)
				err := elPtrValue.Interface().(Valuer).JSONValue(nestedValues.Path(), val, dfield.Options)
				if err != nil {
					errorsInSlice = append(errorsInSlice, err)
				} else {
					errorsInSlice = append(errorsInSlice, nil)

					if dfield.elemKind == reflect.Ptr {
						sliceValue = reflect.Append(sliceValue, elPtrValue)
					} else {
						sliceValue = reflect.Append(sliceValue, reflect.Indirect(elPtrValue))
					}
				}
			}

			fieldValue.Set(sliceValue)
			if errorsInSlice.Len() > 0 {
				errs = addError(errs, metaName, errorsInSlice)
			}
		case categorySliceOfStructs:
			sliceValue := fieldValue
			var errorsInSlice ErrorSlice

			var i int
			sliceSrc := src.Get(metaName)
			for ; true; i += 1 {
				nestedValues := sliceSrc.Get(fmt.Sprint(i)) // foo_bar.0, foo_bar.1, ...
				if nestedValues.Malformed() {
					return ErrorHash{
						"error": ErrMalformed,
					}
				}

				if nestedValues.Empty() {
					break
				}
				elPtrValue := reflect.New(dfield.elemIndirectedType)

				if err := dfield.StructDecoder.decode(elPtrValue, nestedValues); err != nil {
					errorsInSlice = append(errorsInSlice, err)
				} else {
					errorsInSlice = append(errorsInSlice, nil)

					if dfield.elemKind == reflect.Ptr {
						sliceValue = reflect.Append(sliceValue, elPtrValue)
					} else {
						sliceValue = reflect.Append(sliceValue, reflect.Indirect(elPtrValue))
					}
				}
			}

			// Validate the length of the slice
			if dfield.MinLengthPresent && dfield.MinLength > i {
				errs = addError(errs, metaName, ErrMinLength)
			} else if dfield.MaxLengthPresent && dfield.MaxLength < i {
				errs = addError(errs, metaName, ErrMaxLength)
			} else {
				fieldValue.Set(sliceValue)
				if errorsInSlice.Len() > 0 {
					errs = addError(errs, metaName, errorsInSlice)
				}
			}
		case categoryAllFieldsMap:
			fieldValue.Set(reflect.ValueOf(src.ValueMap()))
		}
	}

	return errs
}

// Given the decoder, makes a new struct and tries to map the values onto it. If it succeeds, returns that struct. Otherwise, returns the errors.
func (d *Decoder) NewDecodedValues(values url.Values) (interface{}, ErrorHash) {
	return d.NewDecoded(values, nil)
}

// NewDecoded empties io.Reader and uses its []byte to create json source.
//
// It is often common to call req.ParseForm() before calling this function to obtain url.Values from http request.
// Although ParseForm also reads http request body, it will only do so if the content type is either
// "application/x-www-form-urlencoded" or "multipart/form-data". Therefore, in this case, this function can
// handle both json and form-encoded input.
func (d *Decoder) NewDecoded(values url.Values, r io.Reader) (interface{}, ErrorHash) {
	var b []byte
	if r != nil {
		var err error
		b, err = ioutil.ReadAll(r)
		if err != nil {
			// error hash is used to make it compatible with DecodeValues.
			return nil, NewHash("error", err.Error())
		}
	}

	dest := reflect.New(d.StructType).Interface()
	if err := d.Decode(dest, values, b); err != nil {
		return nil, err
	}
	return dest, nil
}

func addError(errs ErrorHash, key string, value Errorable) ErrorHash {
	if errs == nil {
		errs = make(ErrorHash)
	}
	errs[key] = value
	return errs
}
