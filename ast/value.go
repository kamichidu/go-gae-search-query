package ast

import (
	"encoding/json"
	"time"
)

type Value interface {
	isValue()
}

type TimeValue time.Time

func (v TimeValue) isValue() {}

func (v *TimeValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"T": (*time.Time)(v),
	})
}

type FloatValue float64

func (v FloatValue) isValue() {}

func (v FloatValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"F": float64(v),
	})
}

type IntegerValue int64

func (v IntegerValue) isValue() {}

func (v IntegerValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"I": int64(v),
	})
}

type BoolValue bool

func (v BoolValue) isValue() {}

func (v BoolValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"B": bool(v),
	})
}

type StringValue string

func (v StringValue) isValue() {}

func (v StringValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"S": string(v),
	})
}
