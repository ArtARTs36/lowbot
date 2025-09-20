package callback

type PassEnumValue struct {
	Value string `json:"value"`
}

func NewEnum(value string) *Callback {
	return NewCallback(TypeEnum, &PassEnumValue{Value: value})
}

func (v PassEnumValue) value() {}
