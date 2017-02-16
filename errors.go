package meta

import (
	"encoding/json"
)

type Errorable interface {
	// Errorable should be a go error
	error
	// ErrorKind is a useless method, and probably a poor design of this. My goal is a recursive Errorable type which is an ErrorHash OR ErrorAtom OR ErrorSlice
	ErrorKind() string
}

// assert Errorable at compile-time
var (
	_ Errorable = ErrorHash(nil)
	_ Errorable = ErrorAtom("")
	_ Errorable = ErrorSlice(nil)
)

type ErrorHash map[string]Errorable
type ErrorAtom string
type ErrorSlice []Errorable

func (e ErrorAtom) ErrorKind() string {
	return string(e)
}

func (eh ErrorHash) ErrorKind() string {
	return "invalid_hash"
}

func (ea ErrorSlice) ErrorKind() string {
	return "invalid_slice"
}

func (e ErrorAtom) Error() string {
	return string(e)
}

func (eh ErrorHash) Error() string {
	j, _ := json.Marshal(eh)
	return string(j)
}

func (ea ErrorSlice) Error() string {
	return "invalid_slice"
}

// Len returns the number of non-nil error
func (ea ErrorSlice) Len() (n int) {
	for _, e := range ea {
		if e != nil {
			n++
		}
	}
	return
}

// NewHash returns a new hash with a single key/value in it, eg {"error": "too_big"}
func NewHash(key, value string) ErrorHash {
	return ErrorHash{key: ErrorAtom(value)}
}

var (
	ErrMalformed  = ErrorAtom("malformed_json")
	ErrBlank      = ErrorAtom("blank")
	ErrRequired   = ErrorAtom("required")
	ErrMinRunes   = ErrorAtom("min_runes")
	ErrMaxRunes   = ErrorAtom("max_runes")
	ErrMaxBytes   = ErrorAtom("max_bytes")
	ErrUtf8       = ErrorAtom("utf8")
	ErrBool       = ErrorAtom("bool")
	ErrTime       = ErrorAtom("time")
	ErrInt        = ErrorAtom("int")
	ErrIntRange   = ErrorAtom("int_range")
	ErrString     = ErrorAtom("string")
	ErrFloat      = ErrorAtom("float")
	ErrFloatRange = ErrorAtom("float_range")
	ErrMin        = ErrorAtom("min")
	ErrMax        = ErrorAtom("max")
	ErrIn         = ErrorAtom("in")
	ErrMinLength  = ErrorAtom("min_length")
	ErrMaxLength  = ErrorAtom("max_length")
)
