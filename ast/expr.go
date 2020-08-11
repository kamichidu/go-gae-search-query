package ast

import "encoding/json"

type Expr interface {
	isExpr()
}

type And []Expr

func (v And) isExpr() {}

func (v And) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"and": []Expr(v),
	})
}

type Or []Expr

func (v Or) isExpr() {}

func (v Or) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"or": []Expr(v),
	})
}

type Not struct {
	Expr Expr
}

func (v *Not) isExpr() {}

func (v *Not) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"not": v.Expr,
	})
}

type OperatorExpr struct {
	Property string

	Operator Op

	Value Value
}

func (v *OperatorExpr) isExpr() {}

func (v *OperatorExpr) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		v.Operator.String(): map[string]interface{}{
			"property": v.Property,
			"value":    v.Value,
		},
	})
}

type ColonExpr struct {
	Property string

	Expr Expr
}

func (v *ColonExpr) isExpr() {}

func (v *ColonExpr) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		":": map[string]interface{}{
			"property": v.Property,
			"expr":     v.Expr,
		},
	})
}

type KeywordExpr struct {
	Value Value
}

func (v *KeywordExpr) isExpr() {}

func (v *KeywordExpr) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"keyword": map[string]interface{}{
			"value": v.Value,
		},
	})
}
