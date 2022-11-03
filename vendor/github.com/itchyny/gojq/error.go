package gojq

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// ValueError is an interface for errors with a value for internal function.
// Return an error implementing this interface when you want to catch error
// values (not error messages) by try-catch, just like built-in error function.
// Refer to WithFunction to add a custom internal function.
type ValueError interface {
	error
	Value() interface{}
}

type expectedObjectError struct {
	v interface{}
}

func (err *expectedObjectError) Error() string {
	return "expected an object but got: " + typeErrorPreview(err.v)
}

type expectedArrayError struct {
	v interface{}
}

func (err *expectedArrayError) Error() string {
	return "expected an array but got: " + typeErrorPreview(err.v)
}

type expectedStringError struct {
	v interface{}
}

func (err *expectedStringError) Error() string {
	return "expected a string but got: " + typeErrorPreview(err.v)
}

type iteratorError struct {
	v interface{}
}

func (err *iteratorError) Error() string {
	return "cannot iterate over: " + typeErrorPreview(err.v)
}

type arrayIndexTooLargeError struct {
	v interface{}
}

func (err *arrayIndexTooLargeError) Error() string {
	return "array index too large: " + previewValue(err.v)
}

type objectKeyNotStringError struct {
	v interface{}
}

func (err *objectKeyNotStringError) Error() string {
	return "expected a string for object key but got: " + typeErrorPreview(err.v)
}

type arrayIndexNotNumberError struct {
	v interface{}
}

func (err *arrayIndexNotNumberError) Error() string {
	return "expected a number for indexing an array but got: " + typeErrorPreview(err.v)
}

type expectedStartEndError struct {
	v interface{}
}

func (err *expectedStartEndError) Error() string {
	return `expected "start" and "end" for slicing but got: ` + typeErrorPreview(err.v)
}

type lengthMismatchError struct {
	name string
	v, x []interface{}
}

func (err *lengthMismatchError) Error() string {
	return "length mismatch in " + err.name + ": " + typeErrorPreview(err.v) + ", " + typeErrorPreview(err.x)
}

type inputNotAllowedError struct{}

func (*inputNotAllowedError) Error() string {
	return "input(s)/0 is not allowed"
}

type funcNotFoundError struct {
	f *Func
}

func (err *funcNotFoundError) Error() string {
	return "function not defined: " + err.f.Name + "/" + strconv.Itoa(len(err.f.Args))
}

type funcTypeError struct {
	name string
	v    interface{}
}

func (err *funcTypeError) Error() string {
	return err.name + " cannot be applied to: " + typeErrorPreview(err.v)
}

type exitCodeError struct {
	value interface{}
	code  int
	halt  bool
}

func (err *exitCodeError) Error() string {
	if s, ok := err.value.(string); ok {
		return "error: " + s
	}
	return "error: " + jsonMarshal(err.value)
}

func (err *exitCodeError) IsEmptyError() bool {
	return err.value == nil
}

func (err *exitCodeError) Value() interface{} {
	return err.value
}

func (err *exitCodeError) ExitCode() int {
	return err.code
}

func (err *exitCodeError) IsHaltError() bool {
	return err.halt
}

type funcContainsError struct {
	l, r interface{}
}

func (err *funcContainsError) Error() string {
	return "cannot check contains(" + previewValue(err.r) + "): " + typeErrorPreview(err.l)
}

type hasKeyTypeError struct {
	l, r interface{}
}

func (err *hasKeyTypeError) Error() string {
	return "cannot check whether " + typeErrorPreview(err.l) + " has a key: " + typeErrorPreview(err.r)
}

type unaryTypeError struct {
	name string
	v    interface{}
}

func (err *unaryTypeError) Error() string {
	return "cannot " + err.name + ": " + typeErrorPreview(err.v)
}

type binopTypeError struct {
	name string
	l, r interface{}
}

func (err *binopTypeError) Error() string {
	return "cannot " + err.name + ": " + typeErrorPreview(err.l) + " and " + typeErrorPreview(err.r)
}

type zeroDivisionError struct {
	l, r interface{}
}

func (err *zeroDivisionError) Error() string {
	return "cannot divide " + typeErrorPreview(err.l) + " by: " + typeErrorPreview(err.r)
}

type zeroModuloError struct {
	l, r interface{}
}

func (err *zeroModuloError) Error() string {
	return "cannot modulo " + typeErrorPreview(err.l) + " by: " + typeErrorPreview(err.r) + ""
}

type formatNotFoundError struct {
	n string
}

func (err *formatNotFoundError) Error() string {
	return "format not defined: " + err.n
}

type formatCsvTsvRowError struct {
	typ string
	v   interface{}
}

func (err *formatCsvTsvRowError) Error() string {
	return "invalid " + err.typ + " row: " + typeErrorPreview(err.v)
}

type formatShError struct {
	v interface{}
}

func (err *formatShError) Error() string {
	return "cannot escape for shell: " + typeErrorPreview(err.v)
}

type tooManyVariableValuesError struct{}

func (err *tooManyVariableValuesError) Error() string {
	return "too many variable values provided"
}

type expectedVariableError struct {
	n string
}

func (err *expectedVariableError) Error() string {
	return "variable defined but not bound: " + err.n
}

type variableNotFoundError struct {
	n string
}

func (err *variableNotFoundError) Error() string {
	return "variable not defined: " + err.n
}

type variableNameError struct {
	n string
}

func (err *variableNameError) Error() string {
	return "invalid variable name: " + err.n
}

type breakError struct {
	n string
	v interface{}
}

func (err *breakError) Error() string {
	return "label not defined: " + err.n
}

func (err *breakError) ExitCode() int {
	return 3
}

type tryEndError struct {
	err error
}

func (err *tryEndError) Error() string {
	return err.err.Error()
}

type invalidPathError struct {
	v interface{}
}

func (err *invalidPathError) Error() string {
	return "invalid path against: " + typeErrorPreview(err.v)
}

type invalidPathIterError struct {
	v interface{}
}

func (err *invalidPathIterError) Error() string {
	return "invalid path on iterating against: " + typeErrorPreview(err.v)
}

type getpathError struct {
	v, path interface{}
}

func (err *getpathError) Error() string {
	return "cannot getpath with " + previewValue(err.path) + " against: " + typeErrorPreview(err.v) + ""
}

type queryParseError struct {
	typ, fname, contents string
	err                  error
}

func (err *queryParseError) QueryParseError() (string, string, string, error) {
	return err.typ, err.fname, err.contents, err.err
}

func (err *queryParseError) Error() string {
	return "invalid " + err.typ + ": " + err.fname + ": " + err.err.Error()
}

type jsonParseError struct {
	fname, contents string
	err             error
}

func (err *jsonParseError) JSONParseError() (string, string, error) {
	return err.fname, err.contents, err.err
}

func (err *jsonParseError) Error() string {
	return "invalid json: " + err.fname + ": " + err.err.Error()
}

func typeErrorPreview(v interface{}) string {
	if _, ok := v.(Iter); ok {
		return "gojq.Iter"
	}
	p := preview(v)
	if p != "" {
		p = " (" + p + ")"
	}
	return typeof(v) + p
}

func typeof(v interface{}) string {
	switch v := v.(type) {
	case nil:
		return "null"
	case bool:
		return "boolean"
	case int, float64, *big.Int:
		return "number"
	case string:
		return "string"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		panic(fmt.Sprintf("invalid value: %v", v))
	}
}

type limitedWriter struct {
	buf []byte
	off int
}

func (w *limitedWriter) Write(bs []byte) (int, error) {
	n := copy(w.buf[w.off:], bs)
	if w.off += n; w.off == len(w.buf) {
		panic(nil)
	}
	return n, nil
}

func (w *limitedWriter) WriteByte(b byte) error {
	w.buf[w.off] = b
	if w.off++; w.off == len(w.buf) {
		panic(nil)
	}
	return nil
}

func (w *limitedWriter) WriteString(s string) (int, error) {
	n := copy(w.buf[w.off:], s)
	if w.off += n; w.off == len(w.buf) {
		panic(nil)
	}
	return n, nil
}

func (w *limitedWriter) String() string {
	return string(w.buf[:w.off])
}

func jsonLimitedMarshal(v interface{}, n int) (s string) {
	w := &limitedWriter{buf: make([]byte, n)}
	defer func() {
		recover()
		s = w.String()
	}()
	(&encoder{w: w}).encode(v)
	return
}

func preview(v interface{}) string {
	if v == nil {
		return ""
	}
	s := jsonLimitedMarshal(v, 32)
	if l := 30; len(s) > l {
		var trailing string
		switch v.(type) {
		case string:
			trailing = ` ..."`
		case []interface{}:
			trailing = " ...]"
		case map[string]interface{}:
			trailing = " ...}"
		default:
			trailing = " ..."
		}
		var sb strings.Builder
		sb.Grow(l + 5)
		for _, c := range s {
			sb.WriteRune(c)
			if sb.Len() >= l-len(trailing) {
				sb.WriteString(trailing)
				break
			}
		}
		s = sb.String()
	}
	return s
}

func previewValue(v interface{}) string {
	if v == nil {
		return "null"
	}
	return preview(v)
}
