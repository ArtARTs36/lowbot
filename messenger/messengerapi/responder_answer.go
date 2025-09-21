package messengerapi

type Answer struct {
	Text    string
	Menu    []string
	Enum    Enum
	Buttons []Button
}

type Enum struct {
	Values []EnumItem
}

type EnumItem struct {
	Value string
	Title string
}

func (e *Enum) Valid() bool {
	return len(e.Values) > 0
}

func EnumFromList(values []string) Enum {
	en := Enum{
		Values: make([]EnumItem, len(values)),
	}

	for i, v := range values {
		en.Values[i] = EnumItem{
			Value: v,
			Title: v,
		}
	}

	return en
}

func EnumFromMap(values map[string]string) Enum {
	en := Enum{
		Values: make([]EnumItem, len(values)),
	}

	i := 0
	for value, title := range values {
		en.Values[i] = EnumItem{
			Value: value,
			Title: title,
		}
		i++
	}

	return en
}
