package gojq

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
)

// Query represents the abstract syntax tree of a jq query.
type Query struct {
	Meta     *ConstObject
	Imports  []*Import
	FuncDefs []*FuncDef
	Term     *Term
	Left     *Query
	Op       Operator
	Right    *Query
	Func     string
}

// Run the query.
//
// It is safe to call this method of a *Query in multiple goroutines.
func (e *Query) Run(v interface{}) Iter {
	return e.RunWithContext(context.Background(), v)
}

// RunWithContext runs the query with context.
func (e *Query) RunWithContext(ctx context.Context, v interface{}) Iter {
	code, err := Compile(e)
	if err != nil {
		return NewIter(err)
	}
	return code.RunWithContext(ctx, v)
}

func (e *Query) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Query) writeTo(s *strings.Builder) {
	if e.Meta != nil {
		s.WriteString("module ")
		e.Meta.writeTo(s)
		s.WriteString(";\n")
	}
	for _, im := range e.Imports {
		im.writeTo(s)
	}
	for i, fd := range e.FuncDefs {
		if i > 0 {
			s.WriteByte(' ')
		}
		fd.writeTo(s)
	}
	if len(e.FuncDefs) > 0 {
		s.WriteByte(' ')
	}
	if e.Func != "" {
		s.WriteString(e.Func)
	} else if e.Term != nil {
		e.Term.writeTo(s)
	} else if e.Right != nil {
		e.Left.writeTo(s)
		if e.Op == OpComma {
			s.WriteString(", ")
		} else {
			s.WriteByte(' ')
			s.WriteString(e.Op.String())
			s.WriteByte(' ')
		}
		e.Right.writeTo(s)
	}
}

func (e *Query) minify() {
	for _, e := range e.FuncDefs {
		e.Minify()
	}
	if e.Term != nil {
		if name := e.Term.toFunc(); name != "" {
			e.Term = nil
			e.Func = name
		} else {
			e.Term.minify()
		}
	} else if e.Right != nil {
		e.Left.minify()
		e.Right.minify()
	}
}

func (e *Query) toIndices() []interface{} {
	if e.FuncDefs != nil || e.Right != nil || e.Term == nil {
		return nil
	}
	return e.Term.toIndices()
}

// Import ...
type Import struct {
	ImportPath  string
	ImportAlias string
	IncludePath string
	Meta        *ConstObject
}

func (e *Import) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Import) writeTo(s *strings.Builder) {
	if e.ImportPath != "" {
		s.WriteString("import ")
		s.WriteString(strconv.Quote(e.ImportPath))
		s.WriteString(" as ")
		s.WriteString(e.ImportAlias)
	} else {
		s.WriteString("include ")
		s.WriteString(strconv.Quote(e.IncludePath))
	}
	if e.Meta != nil {
		s.WriteByte(' ')
		e.Meta.writeTo(s)
	}
	s.WriteString(";\n")
}

// FuncDef ...
type FuncDef struct {
	Name string
	Args []string
	Body *Query
}

func (e *FuncDef) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *FuncDef) writeTo(s *strings.Builder) {
	s.WriteString("def ")
	s.WriteString(e.Name)
	if len(e.Args) > 0 {
		s.WriteByte('(')
		for i, e := range e.Args {
			if i > 0 {
				s.WriteString("; ")
			}
			s.WriteString(e)
		}
		s.WriteByte(')')
	}
	s.WriteString(": ")
	e.Body.writeTo(s)
	s.WriteByte(';')
}

// Minify ...
func (e *FuncDef) Minify() {
	e.Body.minify()
}

// Term ...
type Term struct {
	Type       TermType
	Index      *Index
	Func       *Func
	Object     *Object
	Array      *Array
	Number     string
	Unary      *Unary
	Format     string
	Str        *String
	If         *If
	Try        *Try
	Reduce     *Reduce
	Foreach    *Foreach
	Label      *Label
	Break      string
	Query      *Query
	SuffixList []*Suffix
}

func (e *Term) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Term) writeTo(s *strings.Builder) {
	switch e.Type {
	case TermTypeIdentity:
		s.WriteByte('.')
	case TermTypeRecurse:
		s.WriteString("..")
	case TermTypeNull:
		s.WriteString("null")
	case TermTypeTrue:
		s.WriteString("true")
	case TermTypeFalse:
		s.WriteString("false")
	case TermTypeIndex:
		e.Index.writeTo(s)
	case TermTypeFunc:
		e.Func.writeTo(s)
	case TermTypeObject:
		e.Object.writeTo(s)
	case TermTypeArray:
		e.Array.writeTo(s)
	case TermTypeNumber:
		s.WriteString(e.Number)
	case TermTypeUnary:
		e.Unary.writeTo(s)
	case TermTypeFormat:
		s.WriteString(e.Format)
		if e.Str != nil {
			s.WriteByte(' ')
			e.Str.writeTo(s)
		}
	case TermTypeString:
		e.Str.writeTo(s)
	case TermTypeIf:
		e.If.writeTo(s)
	case TermTypeTry:
		e.Try.writeTo(s)
	case TermTypeReduce:
		e.Reduce.writeTo(s)
	case TermTypeForeach:
		e.Foreach.writeTo(s)
	case TermTypeLabel:
		e.Label.writeTo(s)
	case TermTypeBreak:
		s.WriteString("break ")
		s.WriteString(e.Break)
	case TermTypeQuery:
		s.WriteByte('(')
		e.Query.writeTo(s)
		s.WriteByte(')')
	}
	for _, e := range e.SuffixList {
		e.writeTo(s)
	}
}

func (e *Term) minify() {
	switch e.Type {
	case TermTypeIndex:
		e.Index.minify()
	case TermTypeFunc:
		e.Func.minify()
	case TermTypeObject:
		e.Object.minify()
	case TermTypeArray:
		e.Array.minify()
	case TermTypeUnary:
		e.Unary.minify()
	case TermTypeFormat:
		if e.Str != nil {
			e.Str.minify()
		}
	case TermTypeString:
		e.Str.minify()
	case TermTypeIf:
		e.If.minify()
	case TermTypeTry:
		e.Try.minify()
	case TermTypeReduce:
		e.Reduce.minify()
	case TermTypeForeach:
		e.Foreach.minify()
	case TermTypeLabel:
		e.Label.minify()
	case TermTypeQuery:
		e.Query.minify()
	}
	for _, e := range e.SuffixList {
		e.minify()
	}
}

func (e *Term) toFunc() string {
	if len(e.SuffixList) != 0 {
		return ""
	}
	// ref: compiler#compileQuery
	switch e.Type {
	case TermTypeIdentity:
		return "."
	case TermTypeRecurse:
		return ".."
	case TermTypeNull:
		return "null"
	case TermTypeTrue:
		return "true"
	case TermTypeFalse:
		return "false"
	case TermTypeFunc:
		return e.Func.toFunc()
	default:
		return ""
	}
}

func (e *Term) toIndices() []interface{} {
	if e.Index != nil {
		xs := e.Index.toIndices()
		if xs == nil {
			return nil
		}
		for _, s := range e.SuffixList {
			x := s.toIndices()
			if x == nil {
				return nil
			}
			xs = append(xs, x...)
		}
		return xs
	} else if e.Query != nil && len(e.SuffixList) == 0 {
		return e.Query.toIndices()
	} else {
		return nil
	}
}

// Unary ...
type Unary struct {
	Op   Operator
	Term *Term
}

func (e *Unary) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Unary) writeTo(s *strings.Builder) {
	s.WriteString(e.Op.String())
	e.Term.writeTo(s)
}

func (e *Unary) minify() {
	e.Term.minify()
}

// Pattern ...
type Pattern struct {
	Name   string
	Array  []*Pattern
	Object []*PatternObject
}

func (e *Pattern) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Pattern) writeTo(s *strings.Builder) {
	if e.Name != "" {
		s.WriteString(e.Name)
	} else if len(e.Array) > 0 {
		s.WriteByte('[')
		for i, e := range e.Array {
			if i > 0 {
				s.WriteString(", ")
			}
			e.writeTo(s)
		}
		s.WriteByte(']')
	} else if len(e.Object) > 0 {
		s.WriteByte('{')
		for i, e := range e.Object {
			if i > 0 {
				s.WriteString(", ")
			}
			e.writeTo(s)
		}
		s.WriteByte('}')
	}
}

// PatternObject ...
type PatternObject struct {
	Key       string
	KeyString *String
	KeyQuery  *Query
	Val       *Pattern
	KeyOnly   string
}

func (e *PatternObject) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *PatternObject) writeTo(s *strings.Builder) {
	if e.Key != "" {
		s.WriteString(e.Key)
	} else if e.KeyString != nil {
		e.KeyString.writeTo(s)
	} else if e.KeyQuery != nil {
		s.WriteByte('(')
		e.KeyQuery.writeTo(s)
		s.WriteByte(')')
	}
	if e.Val != nil {
		s.WriteString(": ")
		e.Val.writeTo(s)
	}
	if e.KeyOnly != "" {
		s.WriteString(e.KeyOnly)
	}
}

// Index ...
type Index struct {
	Name    string
	Str     *String
	Start   *Query
	End     *Query
	IsSlice bool
}

func (e *Index) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Index) writeTo(s *strings.Builder) {
	if l := s.Len(); l > 0 {
		// ". .x" != "..x" and "0 .x" != "0.x"
		if c := s.String()[l-1]; c == '.' || '0' <= c && c <= '9' {
			s.WriteByte(' ')
		}
	}
	s.WriteByte('.')
	e.writeSuffixTo(s)
}

func (e *Index) writeSuffixTo(s *strings.Builder) {
	if e.Name != "" {
		s.WriteString(e.Name)
	} else {
		if e.Str != nil {
			e.Str.writeTo(s)
		} else {
			s.WriteByte('[')
			if e.IsSlice {
				if e.Start != nil {
					e.Start.writeTo(s)
				}
				s.WriteByte(':')
				if e.End != nil {
					e.End.writeTo(s)
				}
			} else {
				e.Start.writeTo(s)
			}
			s.WriteByte(']')
		}
	}
}

func (e *Index) minify() {
	if e.Str != nil {
		e.Str.minify()
	}
	if e.Start != nil {
		e.Start.minify()
	}
	if e.End != nil {
		e.End.minify()
	}
}

func (e *Index) toIndices() []interface{} {
	if e.Name == "" {
		return nil
	}
	return []interface{}{e.Name}
}

// Func ...
type Func struct {
	Name string
	Args []*Query
}

func (e *Func) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Func) writeTo(s *strings.Builder) {
	s.WriteString(e.Name)
	if len(e.Args) > 0 {
		s.WriteByte('(')
		for i, e := range e.Args {
			if i > 0 {
				s.WriteString("; ")
			}
			e.writeTo(s)
		}
		s.WriteByte(')')
	}
}

func (e *Func) minify() {
	for _, x := range e.Args {
		x.minify()
	}
}

func (e *Func) toFunc() string {
	if len(e.Args) != 0 {
		return ""
	}
	return e.Name
}

// String ...
type String struct {
	Str     string
	Queries []*Query
}

func (e *String) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *String) writeTo(s *strings.Builder) {
	if e.Queries == nil {
		s.WriteString(strconv.Quote(e.Str))
		return
	}
	s.WriteByte('"')
	for _, e := range e.Queries {
		if e.Term.Str == nil {
			s.WriteString(`\`)
			e.writeTo(s)
		} else {
			es := e.String()
			s.WriteString(es[1 : len(es)-1])
		}
	}
	s.WriteByte('"')
}

func (e *String) minify() {
	for _, e := range e.Queries {
		e.minify()
	}
}

// Object ...
type Object struct {
	KeyVals []*ObjectKeyVal
}

func (e *Object) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Object) writeTo(s *strings.Builder) {
	if len(e.KeyVals) == 0 {
		s.WriteString("{}")
		return
	}
	s.WriteString("{ ")
	for i, kv := range e.KeyVals {
		if i > 0 {
			s.WriteString(", ")
		}
		kv.writeTo(s)
	}
	s.WriteString(" }")
}

func (e *Object) minify() {
	for _, e := range e.KeyVals {
		e.minify()
	}
}

// ObjectKeyVal ...
type ObjectKeyVal struct {
	Key           string
	KeyString     *String
	KeyQuery      *Query
	Val           *ObjectVal
	KeyOnly       string
	KeyOnlyString *String
}

func (e *ObjectKeyVal) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *ObjectKeyVal) writeTo(s *strings.Builder) {
	if e.Key != "" {
		s.WriteString(e.Key)
	} else if e.KeyString != nil {
		e.KeyString.writeTo(s)
	} else if e.KeyQuery != nil {
		s.WriteByte('(')
		e.KeyQuery.writeTo(s)
		s.WriteByte(')')
	}
	if e.Val != nil {
		s.WriteString(": ")
		e.Val.writeTo(s)
	}
	if e.KeyOnly != "" {
		s.WriteString(e.KeyOnly)
	} else if e.KeyOnlyString != nil {
		e.KeyOnlyString.writeTo(s)
	}
}

func (e *ObjectKeyVal) minify() {
	if e.KeyString != nil {
		e.KeyString.minify()
	} else if e.KeyQuery != nil {
		e.KeyQuery.minify()
	}
	if e.Val != nil {
		e.Val.minify()
	}
	if e.KeyOnlyString != nil {
		e.KeyOnlyString.minify()
	}
}

// ObjectVal ...
type ObjectVal struct {
	Queries []*Query
}

func (e *ObjectVal) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *ObjectVal) writeTo(s *strings.Builder) {
	for i, e := range e.Queries {
		if i > 0 {
			s.WriteString(" | ")
		}
		e.writeTo(s)
	}
}

func (e *ObjectVal) minify() {
	for _, e := range e.Queries {
		e.minify()
	}
}

// Array ...
type Array struct {
	Query *Query
}

func (e *Array) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Array) writeTo(s *strings.Builder) {
	s.WriteByte('[')
	if e.Query != nil {
		e.Query.writeTo(s)
	}
	s.WriteByte(']')
}

func (e *Array) minify() {
	if e.Query != nil {
		e.Query.minify()
	}
}

// Suffix ...
type Suffix struct {
	Index    *Index
	Iter     bool
	Optional bool
	Bind     *Bind
}

func (e *Suffix) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Suffix) writeTo(s *strings.Builder) {
	if e.Index != nil {
		if e.Index.Name != "" || e.Index.Str != nil {
			e.Index.writeTo(s)
		} else {
			e.Index.writeSuffixTo(s)
		}
	} else if e.Iter {
		s.WriteString("[]")
	} else if e.Optional {
		s.WriteByte('?')
	} else if e.Bind != nil {
		e.Bind.writeTo(s)
	}
}

func (e *Suffix) minify() {
	if e.Index != nil {
		e.Index.minify()
	} else if e.Bind != nil {
		e.Bind.minify()
	}
}

func (e *Suffix) toTerm() (*Term, bool) {
	if e.Index != nil {
		return &Term{Type: TermTypeIndex, Index: e.Index}, true
	} else if e.Iter {
		return &Term{Type: TermTypeIdentity, SuffixList: []*Suffix{{Iter: true}}}, true
	} else {
		return nil, false
	}
}

func (e *Suffix) toIndices() []interface{} {
	if e.Index == nil {
		return nil
	}
	return e.Index.toIndices()
}

// Bind ...
type Bind struct {
	Patterns []*Pattern
	Body     *Query
}

func (e *Bind) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Bind) writeTo(s *strings.Builder) {
	for i, p := range e.Patterns {
		if i == 0 {
			s.WriteString(" as ")
			p.writeTo(s)
			s.WriteByte(' ')
		} else {
			s.WriteString("?// ")
			p.writeTo(s)
			s.WriteByte(' ')
		}
	}
	s.WriteString("| ")
	e.Body.writeTo(s)
}

func (e *Bind) minify() {
	e.Body.minify()
}

// If ...
type If struct {
	Cond *Query
	Then *Query
	Elif []*IfElif
	Else *Query
}

func (e *If) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *If) writeTo(s *strings.Builder) {
	s.WriteString("if ")
	e.Cond.writeTo(s)
	s.WriteString(" then ")
	e.Then.writeTo(s)
	for _, e := range e.Elif {
		s.WriteByte(' ')
		e.writeTo(s)
	}
	if e.Else != nil {
		s.WriteString(" else ")
		e.Else.writeTo(s)
	}
	s.WriteString(" end")
}

func (e *If) minify() {
	e.Cond.minify()
	e.Then.minify()
	for _, x := range e.Elif {
		x.minify()
	}
	if e.Else != nil {
		e.Else.minify()
	}
}

// IfElif ...
type IfElif struct {
	Cond *Query
	Then *Query
}

func (e *IfElif) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *IfElif) writeTo(s *strings.Builder) {
	s.WriteString("elif ")
	e.Cond.writeTo(s)
	s.WriteString(" then ")
	e.Then.writeTo(s)
}

func (e *IfElif) minify() {
	e.Cond.minify()
	e.Then.minify()
}

// Try ...
type Try struct {
	Body  *Query
	Catch *Query
}

func (e *Try) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Try) writeTo(s *strings.Builder) {
	s.WriteString("try ")
	e.Body.writeTo(s)
	if e.Catch != nil {
		s.WriteString(" catch ")
		e.Catch.writeTo(s)
	}
}

func (e *Try) minify() {
	e.Body.minify()
	if e.Catch != nil {
		e.Catch.minify()
	}
}

// Reduce ...
type Reduce struct {
	Term    *Term
	Pattern *Pattern
	Start   *Query
	Update  *Query
}

func (e *Reduce) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Reduce) writeTo(s *strings.Builder) {
	s.WriteString("reduce ")
	e.Term.writeTo(s)
	s.WriteString(" as ")
	e.Pattern.writeTo(s)
	s.WriteString(" (")
	e.Start.writeTo(s)
	s.WriteString("; ")
	e.Update.writeTo(s)
	s.WriteByte(')')
}

func (e *Reduce) minify() {
	e.Term.minify()
	e.Start.minify()
	e.Update.minify()
}

// Foreach ...
type Foreach struct {
	Term    *Term
	Pattern *Pattern
	Start   *Query
	Update  *Query
	Extract *Query
}

func (e *Foreach) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Foreach) writeTo(s *strings.Builder) {
	s.WriteString("foreach ")
	e.Term.writeTo(s)
	s.WriteString(" as ")
	e.Pattern.writeTo(s)
	s.WriteString(" (")
	e.Start.writeTo(s)
	s.WriteString("; ")
	e.Update.writeTo(s)
	if e.Extract != nil {
		s.WriteString("; ")
		e.Extract.writeTo(s)
	}
	s.WriteByte(')')
}

func (e *Foreach) minify() {
	e.Term.minify()
	e.Start.minify()
	e.Update.minify()
	if e.Extract != nil {
		e.Extract.minify()
	}
}

// Label ...
type Label struct {
	Ident string
	Body  *Query
}

func (e *Label) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Label) writeTo(s *strings.Builder) {
	s.WriteString("label ")
	s.WriteString(e.Ident)
	s.WriteString(" | ")
	e.Body.writeTo(s)
}

func (e *Label) minify() {
	e.Body.minify()
}

// ConstTerm ...
type ConstTerm struct {
	Object *ConstObject
	Array  *ConstArray
	Number string
	Str    string
	Null   bool
	True   bool
	False  bool
}

func (e *ConstTerm) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *ConstTerm) writeTo(s *strings.Builder) {
	if e.Object != nil {
		e.Object.writeTo(s)
	} else if e.Array != nil {
		e.Array.writeTo(s)
	} else if e.Number != "" {
		s.WriteString(e.Number)
	} else if e.Null {
		s.WriteString("null")
	} else if e.True {
		s.WriteString("true")
	} else if e.False {
		s.WriteString("false")
	} else {
		s.WriteString(strconv.Quote(e.Str))
	}
}

func (e *ConstTerm) toValue() interface{} {
	if e.Object != nil {
		return e.Object.ToValue()
	} else if e.Array != nil {
		return e.Array.toValue()
	} else if e.Number != "" {
		return normalizeNumber(json.Number(e.Number))
	} else if e.Null {
		return nil
	} else if e.True {
		return true
	} else if e.False {
		return false
	} else {
		return e.Str
	}
}

// ConstObject ...
type ConstObject struct {
	KeyVals []*ConstObjectKeyVal
}

func (e *ConstObject) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *ConstObject) writeTo(s *strings.Builder) {
	if len(e.KeyVals) == 0 {
		s.WriteString("{}")
		return
	}
	s.WriteString("{ ")
	for i, kv := range e.KeyVals {
		if i > 0 {
			s.WriteString(", ")
		}
		kv.writeTo(s)
	}
	s.WriteString(" }")
}

// ToValue converts the object to map[string]interface{}.
func (e *ConstObject) ToValue() map[string]interface{} {
	if e == nil {
		return nil
	}
	v := make(map[string]interface{}, len(e.KeyVals))
	for _, e := range e.KeyVals {
		key := e.Key
		if key == "" {
			key = e.KeyString
		}
		v[key] = e.Val.toValue()
	}
	return v
}

// ConstObjectKeyVal ...
type ConstObjectKeyVal struct {
	Key       string
	KeyString string
	Val       *ConstTerm
}

func (e *ConstObjectKeyVal) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *ConstObjectKeyVal) writeTo(s *strings.Builder) {
	if e.Key != "" {
		s.WriteString(e.Key)
	} else {
		s.WriteString(e.KeyString)
	}
	s.WriteString(": ")
	e.Val.writeTo(s)
}

// ConstArray ...
type ConstArray struct {
	Elems []*ConstTerm
}

func (e *ConstArray) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *ConstArray) writeTo(s *strings.Builder) {
	s.WriteByte('[')
	for i, e := range e.Elems {
		if i > 0 {
			s.WriteString(", ")
		}
		e.writeTo(s)
	}
	s.WriteByte(']')
}

func (e *ConstArray) toValue() []interface{} {
	v := make([]interface{}, len(e.Elems))
	for i, e := range e.Elems {
		v[i] = e.toValue()
	}
	return v
}
