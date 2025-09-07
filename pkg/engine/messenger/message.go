package messenger

type Message interface {
	GetID() string
	GetChatID() string
	GetBody() string

	ExtractCommandName() string

	Respond(answer *Answer) error
}

type Answer struct {
	Text string
	Menu []string
	Enum Enum
}

type Enum struct {
	Values map[string]string
}

func (e *Enum) Valid() bool {
	return len(e.Values) > 0
}

func EnumFromList(values []string) Enum {
	en := Enum{
		Values: make(map[string]string),
	}

	for _, v := range values {
		en.Values[v] = v
	}

	return en
}
