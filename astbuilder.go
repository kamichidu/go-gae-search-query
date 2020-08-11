package searchquery

import (
	"container/list"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/kamichidu/go-gae-search-query/ast"
)

type astBuilder struct {
	Expr ast.Expr

	Debug bool

	state list.List

	stateStack list.List
}

func (a *astBuilder) pushState(v interface{}) {
	a.state.PushFront(v)
}

func (a *astBuilder) popState() interface{} {
	ele := a.state.Front()
	if ele == nil {
		return nil
	}
	a.state.Remove(ele)
	return ele.Value
}

func (a *astBuilder) peekState() interface{} {
	ele := a.state.Front()
	if ele == nil {
		return nil
	}
	return ele.Value
}

func (a *astBuilder) log(format string, args ...interface{}) {
	if !a.Debug {
		return
	}
	log.Printf("trace: "+format, args...)
}

func (a *astBuilder) finalize() {
	a.log("finalize")

	if n := a.stateStack.Len(); n > 0 {
		panic(fmt.Sprintf("%s: invalid state stack: remaining %d state stacks", pkgName, n))
	}
	if n := a.state.Len(); n != 1 {
		panic(fmt.Sprintf("%s: invalid state: remaining %d state", pkgName, n))
	}
	a.Expr = a.state.Remove(a.state.Front()).(ast.Expr)
}

func (a *astBuilder) reduceAnd() {
	a.log("reduceAnd")

	// remaining an expr, do nothing
	if n := a.state.Len(); n == 1 {
		return
	}
	var and ast.And
	for ele := a.state.Back(); ele != nil; ele = ele.Prev() {
		and = append(and, ele.Value.(ast.Expr))
	}
	a.state.Init()
	a.state.PushFront(and)
}

func (a *astBuilder) pushNewState() {
	a.log("pushNewState")

	a.stateStack.PushFront(a.state)
	a.state.Init()
}

func (a *astBuilder) popNewState() {
	a.log("popNewState")

	if n := a.state.Len(); n != 1 {
		panic(fmt.Sprintf("%s: invalid state: remaining %d states", pkgName, n))
	}
	expr := a.state.Front().Value.(ast.Expr)
	ele := a.stateStack.Front()
	a.stateStack.Remove(ele)
	prevState := ele.Value.(list.List)
	a.state.Init()
	a.state.PushFrontList(&prevState)
	a.state.PushFront(expr)
}

func (a *astBuilder) pushOr() {
	a.log("pushOr")

	expr2_ := a.popState()
	expr1_ := a.popState()

	expr1, ok := expr1_.(ast.Expr)
	if !ok {
		panic(fmt.Sprintf("%s: invalid state: expr1 = %T", pkgName, expr1_))
	}
	expr2, ok := expr2_.(ast.Expr)
	if !ok {
		panic(fmt.Sprintf("%s: invalid state: expr2 = %T", pkgName, expr2_))
	}
	a.pushState(ast.Or{expr1, expr2})
}

func (a *astBuilder) pushNot() {
	a.log("pushNot")

	expr_ := a.popState()

	expr, ok := expr_.(ast.Expr)
	if !ok {
		panic(fmt.Sprintf("%s: invalid state: expr = %T", pkgName, expr_))
	}
	a.pushState(&ast.Not{
		Expr: expr,
	})
}

func (a *astBuilder) pushProperty(s string) {
	a.log("pushProperty %q", s)

	a.pushState(s)
}

func (a *astBuilder) pushOperator(v ast.Op) {
	a.log("pushOperator %v", v)

	a.pushState(v)
}

func (a *astBuilder) pushTimeValue(layout, s string) {
	a.log("pushTimeValue %q %q", layout, s)

	v, err := time.Parse(layout, s)
	if err != nil {
		panic(err)
	}
	a.pushState(ast.TimeValue(v))
}

func (a *astBuilder) pushStringValue(s string) {
	a.log("pushStringValue %q", s)

	a.pushState(ast.StringValue(s))
}

func (a *astBuilder) pushIntegerValue(s string) {
	a.log("pushIntegerValue %q", s)

	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(err)
	}
	a.pushState(ast.IntegerValue(v))
}

func (a *astBuilder) pushFloatValue(s string) {
	a.log("pushFloatValue %q", s)

	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(err)
	}
	a.pushState(ast.FloatValue(v))
}

func (a *astBuilder) pushBoolValue(v bool) {
	a.log("pushBoolValue %v", v)

	a.pushState(ast.BoolValue(v))
}

func (a *astBuilder) pushOperatorExpr() {
	a.log("pushOperatorExpr")

	var expr ast.OperatorExpr
	expr.Value = a.popState().(ast.Value)
	expr.Operator = a.popState().(ast.Op)
	expr.Property = a.popState().(string)
	a.pushState(&expr)
}

func (a *astBuilder) pushColonExpr() {
	a.log("pushColonExpr")

	var expr ast.ColonExpr
	expr.Expr = a.popState().(ast.Expr)
	expr.Property = a.popState().(string)
	a.pushState(&expr)
}

func (a *astBuilder) pushKeywordExpr() {
	a.log("pushKeywordExpr")

	var expr ast.KeywordExpr
	expr.Value = a.popState().(ast.Value)
	a.pushState(&expr)
}
