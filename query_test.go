package searchquery

import (
	"testing"
	"time"

	"github.com/kamichidu/go-gae-search-query/ast"
	"github.com/stretchr/testify/assert"
)

func mustParseTime(s string) time.Time {
	v, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return v
}

func TestParse(t *testing.T) {
	t.Run("", func(t *testing.T) {
		s := `blue`
		expr, err := Parse(s)
		if !assert.NoError(t, err, s) {
			return
		}
		assert.Equal(t, &ast.KeywordExpr{
			Value: ast.StringValue("blue"),
		}, expr, s)
	})
	t.Run("", func(t *testing.T) {
		s := `NOT white`
		expr, err := Parse(s)
		if !assert.NoError(t, err, s) {
			return
		}
		assert.Equal(t, &ast.Not{
			Expr: &ast.KeywordExpr{
				Value: ast.StringValue("white"),
			},
		}, expr, s)
	})
	t.Run("", func(t *testing.T) {
		s := `blue OR red`
		expr, err := Parse(s)
		if !assert.NoError(t, err, s) {
			return
		}
		assert.Equal(t, ast.Or{
			&ast.KeywordExpr{
				Value: ast.StringValue("blue"),
			},
			&ast.KeywordExpr{
				Value: ast.StringValue("red"),
			},
		}, expr, s)
	})
	t.Run("", func(t *testing.T) {
		s := `blue guitar`
		expr, err := Parse(s)
		if !assert.NoError(t, err, s) {
			return
		}
		assert.Equal(t, ast.And{
			&ast.KeywordExpr{
				Value: ast.StringValue("blue"),
			},
			&ast.KeywordExpr{
				Value: ast.StringValue("guitar"),
			},
		}, expr, s)
	})
	t.Run("", func(t *testing.T) {
		s := `model:gibson date < 1965-01-01`
		expr, err := Parse(s)
		if !assert.NoError(t, err, s) {
			return
		}
		assert.Equal(t, ast.And{
			&ast.ColonExpr{
				Property: "model",
				Expr: &ast.KeywordExpr{
					Value: ast.StringValue("gibson"),
				},
			},
			&ast.OperatorExpr{
				Property: "date",
				Operator: ast.OpLt,
				Value:    ast.TimeValue(mustParseTime("1965-01-01T00:00:00Z")),
			},
		}, expr, s)
	})
	t.Run("", func(t *testing.T) {
		s := `title:"Harry Potter" AND pages<500`
		expr, err := Parse(s)
		if !assert.NoError(t, err, s) {
			return
		}
		assert.Equal(t, ast.And{
			&ast.ColonExpr{
				Property: "title",
				Expr: &ast.KeywordExpr{
					Value: ast.StringValue("Harry Potter"),
				},
			},
			&ast.OperatorExpr{
				Property: "pages",
				Operator: ast.OpLt,
				Value:    ast.IntegerValue(500),
			},
		}, expr, s)
	})
	t.Run("", func(t *testing.T) {
		s := `beverage:wine color:(red OR white) NOT country:france`
		expr, err := Parse(s)
		if !assert.NoError(t, err, s) {
			return
		}
		assert.Equal(t, ast.And{
			&ast.ColonExpr{
				Property: "beverage",
				Expr: &ast.KeywordExpr{
					Value: ast.StringValue("wine"),
				},
			},
			&ast.ColonExpr{
				Property: "color",
				Expr: ast.Or{
					&ast.KeywordExpr{
						Value: ast.StringValue("red"),
					},
					&ast.KeywordExpr{
						Value: ast.StringValue("white"),
					},
				},
			},
			&ast.Not{
				Expr: &ast.ColonExpr{
					Property: "country",
					Expr: &ast.KeywordExpr{
						Value: ast.StringValue("france"),
					},
				},
			},
		}, expr, s)
	})
	t.Run("", func(t *testing.T) {
		s := `true false`
		expr, err := Parse(s)
		if !assert.NoError(t, err, s) {
			return
		}
		assert.Equal(t, ast.And{
			&ast.KeywordExpr{
				Value: ast.BoolValue(true),
			},
			&ast.KeywordExpr{
				Value: ast.BoolValue(false),
			},
		}, expr, s)
	})
	t.Run("", func(t *testing.T) {
		s := `NOT cat AND dogs OR horses` // --> (NOT cat) AND (dogs OR horses)
		expr, err := Parse(s)
		if !assert.NoError(t, err, s) {
			return
		}
		assert.Equal(t, ast.And{
			&ast.Not{
				Expr: &ast.KeywordExpr{
					Value: ast.StringValue("cat"),
				},
			},
			ast.Or{
				&ast.KeywordExpr{
					Value: ast.StringValue("dogs"),
				},
				&ast.KeywordExpr{
					Value: ast.StringValue("horses"),
				},
			},
		}, expr, s)
	})
	t.Run("", func(t *testing.T) {
		s := `NOT cat OR dogs AND horses` // --> ((NOT cat) OR dogs) AND horses
		expr, err := Parse(s)
		if !assert.NoError(t, err, s) {
			return
		}
		assert.Equal(t, ast.And{
			ast.Or{
				&ast.Not{
					Expr: &ast.KeywordExpr{
						Value: ast.StringValue("cat"),
					},
				},
				&ast.KeywordExpr{
					Value: ast.StringValue("dogs"),
				},
			},
			&ast.KeywordExpr{
				Value: ast.StringValue("horses"),
			},
		}, expr, s)
	})
	t.Run("", func(t *testing.T) {
		s := `users.user_id = xxx`
		expr, err := Parse(s)
		if !assert.NoError(t, err, s) {
			return
		}
		assert.Equal(t, &ast.OperatorExpr{
			Property: "users.user_id",
			Operator: ast.OpEq,
			Value:    ast.StringValue("xxx"),
		}, expr, s)
	})
}
