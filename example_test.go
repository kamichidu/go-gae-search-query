package searchquery_test

import (
	"fmt"

	sqr "github.com/Masterminds/squirrel"
	searchquery "github.com/kamichidu/go-gae-search-query"
	"github.com/kamichidu/go-gae-search-query/ast"
)

func expr2sqlizer(expr ast.Expr) sqr.Sqlizer {
	switch e := expr.(type) {
	case ast.And:
		var and sqr.And
		for _, v := range e {
			and = append(and, expr2sqlizer(v))
		}
		return and
	case ast.Or:
		var or sqr.Or
		for _, v := range e {
			or = append(or, expr2sqlizer(v))
		}
		return or
	case *ast.Not:
		q, a, err := expr2sqlizer(e.Expr).ToSql()
		if err != nil {
			panic(err)
		}
		return sqr.Expr("NOT ("+q+")", a...)
	case *ast.OperatorExpr:
		switch e.Operator {
		case ast.OpEq:
			return sqr.Eq{
				e.Property: e.Value.Raw(),
			}
		case ast.OpNeq:
			return sqr.NotEq{
				e.Property: e.Value.Raw(),
			}
		case ast.OpLt:
			return sqr.Lt{
				e.Property: e.Value.Raw(),
			}
		case ast.OpLe:
			return sqr.LtOrEq{
				e.Property: e.Value.Raw(),
			}
		case ast.OpGt:
			return sqr.Gt{
				e.Property: e.Value.Raw(),
			}
		case ast.OpGe:
			return sqr.GtOrEq{
				e.Property: e.Value.Raw(),
			}
		default:
			panic(fmt.Sprintf("unknown operator %v", e.Operator))
		}
	case *ast.ColonExpr:
		panic("unsupported `property:expr` syntax")
	case *ast.KeywordExpr:
		panic("unsupported `value` syntax")
	default:
		panic(fmt.Sprintf("unknown expr type %T", expr))
	}
}

func Example() {
	expr, err := searchquery.Parse(`NOT users.name = kamichidu OR users.type != dogs AND users.type = horses`)
	if err != nil {
		panic(err)
	}
	sqlizer := expr2sqlizer(expr)
	q, a, err := sqlizer.ToSql()
	if err != nil {
		panic(err)
	}
	fmt.Printf("QUERY: %s\n", q)
	fmt.Printf("ARGS: %v\n", a)
	// Output:
	// QUERY: ((NOT (users.name = ?) OR users.type <> ?) AND users.type = ?)
	// ARGS: [kamichidu dogs horses]
}
