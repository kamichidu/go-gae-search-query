package ast

import "fmt"

type Op int

func (v Op) String() string {
	switch v {
	case OpEq:
		return "="
	case OpNeq:
		return "!="
	case OpLt:
		return "<"
	case OpLe:
		return "<="
	case OpGt:
		return ">"
	case OpGe:
		return ">="
	default:
		panic(fmt.Sprintf("%s: invalid operator %v", pkgName, v))
	}
}

const (
	op_begin Op = iota
	OpEq
	OpNeq
	OpLt
	OpLe
	OpGt
	OpGe
	op_end
)
