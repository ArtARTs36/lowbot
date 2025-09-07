package callback

import "fmt"

type PassEnumValue struct {
	Value string
}

func NewPassEnumValue(value string) *PassEnumValue {
	return &PassEnumValue{
		Value: value,
	}
}

func (v *PassEnumValue) String() string {
	return fmt.Sprintf("%s:%s", TypePassEnumValue, v.Value)
}
