package searchquery

import (
	"github.com/kamichidu/go-gae-search-query/ast"
)

const (
	pkgName = "searchquery"
)

func Parse(s string) (ast.Expr, error) {
	var q Query
	q.Buffer = s
	q.Init()
	if err := q.Parse(); err != nil {
		return nil, err
	}
	q.Execute()
	// q.PrintSyntaxTree()
	return q.Expr, nil
}
