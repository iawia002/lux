package gojq

// TermType represents the type of Term.
type TermType int

// TermType list.
const (
	TermTypeIdentity TermType = iota + 1
	TermTypeRecurse
	TermTypeNull
	TermTypeTrue
	TermTypeFalse
	TermTypeIndex
	TermTypeFunc
	TermTypeObject
	TermTypeArray
	TermTypeNumber
	TermTypeUnary
	TermTypeFormat
	TermTypeString
	TermTypeIf
	TermTypeTry
	TermTypeReduce
	TermTypeForeach
	TermTypeLabel
	TermTypeBreak
	TermTypeQuery
)

// GoString implements GoStringer.
func (termType TermType) GoString() (str string) {
	defer func() { str = "gojq." + str }()
	switch termType {
	case TermTypeIdentity:
		return "TermTypeIdentity"
	case TermTypeRecurse:
		return "TermTypeRecurse"
	case TermTypeNull:
		return "TermTypeNull"
	case TermTypeTrue:
		return "TermTypeTrue"
	case TermTypeFalse:
		return "TermTypeFalse"
	case TermTypeIndex:
		return "TermTypeIndex"
	case TermTypeFunc:
		return "TermTypeFunc"
	case TermTypeObject:
		return "TermTypeObject"
	case TermTypeArray:
		return "TermTypeArray"
	case TermTypeNumber:
		return "TermTypeNumber"
	case TermTypeUnary:
		return "TermTypeUnary"
	case TermTypeFormat:
		return "TermTypeFormat"
	case TermTypeString:
		return "TermTypeString"
	case TermTypeIf:
		return "TermTypeIf"
	case TermTypeTry:
		return "TermTypeTry"
	case TermTypeReduce:
		return "TermTypeReduce"
	case TermTypeForeach:
		return "TermTypeForeach"
	case TermTypeLabel:
		return "TermTypeLabel"
	case TermTypeBreak:
		return "TermTypeBreak"
	case TermTypeQuery:
		return "TermTypeQuery"
	default:
		panic(termType)
	}
}
