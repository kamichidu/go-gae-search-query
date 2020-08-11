package ast

import (
	"encoding/json"
	"time"
)

type Value interface {
	isValue()

	Raw() interface{}
}

type TimeValue time.Time

func (v TimeValue) isValue() {}

func (v TimeValue) Raw() interface{} {
	return time.Time(v)
}

func (v *TimeValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"T": v.Raw(),
	})
}

type FloatValue float64

func (v FloatValue) isValue() {}

func (v FloatValue) Raw() interface{} {
	return float64(v)
}

func (v FloatValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"F": v.Raw(),
	})
}

type IntegerValue int64

func (v IntegerValue) isValue() {}

func (v IntegerValue) Raw() interface{} {
	return int64(v)
}

func (v IntegerValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"I": v.Raw(),
	})
}

type BoolValue bool

func (v BoolValue) isValue() {}

func (v BoolValue) Raw() interface{} {
	return bool(v)
}

func (v BoolValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"B": v.Raw(),
	})
}

type StringValue string

func (v StringValue) isValue() {}

func (v StringValue) Raw() interface{} {
	return string(v)
}

func (v StringValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"S": v.Raw(),
	})
}
