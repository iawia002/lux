package gojq

type code struct {
	v  interface{}
	op opcode
}

type opcode int

const (
	opnop opcode = iota
	oppush
	oppop
	opdup
	opconst
	opload
	opstore
	opobject
	opappend
	opfork
	opforktrybegin
	opforktryend
	opforkalt
	opforklabel
	opbacktrack
	opjump
	opjumpifnot
	opcall
	opcallrec
	oppushpc
	opcallpc
	opscope
	opret
	opeach
	opexpbegin
	opexpend
	oppathbegin
	oppathend
)

func (op opcode) String() string {
	switch op {
	case opnop:
		return "nop"
	case oppush:
		return "push"
	case oppop:
		return "pop"
	case opdup:
		return "dup"
	case opconst:
		return "const"
	case opload:
		return "load"
	case opstore:
		return "store"
	case opobject:
		return "object"
	case opappend:
		return "append"
	case opfork:
		return "fork"
	case opforktrybegin:
		return "forktrybegin"
	case opforktryend:
		return "forktryend"
	case opforkalt:
		return "forkalt"
	case opforklabel:
		return "forklabel"
	case opbacktrack:
		return "backtrack"
	case opjump:
		return "jump"
	case opjumpifnot:
		return "jumpifnot"
	case opcall:
		return "call"
	case opcallrec:
		return "callrec"
	case oppushpc:
		return "pushpc"
	case opcallpc:
		return "callpc"
	case opscope:
		return "scope"
	case opret:
		return "ret"
	case opeach:
		return "each"
	case opexpbegin:
		return "expbegin"
	case opexpend:
		return "expend"
	case oppathbegin:
		return "pathbegin"
	case oppathend:
		return "pathend"
	default:
		panic(op)
	}
}
