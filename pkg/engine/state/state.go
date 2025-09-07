package state

type State struct {
	chatID      string
	name        string
	commandName string

	transited bool

	data map[string]string
}

func NewState(chatID string, commandName string) *State {
	return &State{
		chatID:      chatID,
		commandName: commandName,
		data:        make(map[string]string),
	}
}

func (m *State) ChatID() string {
	return m.chatID
}

func (m *State) Name() string {
	return m.name
}

func (m *State) CommandName() string {
	return m.commandName
}

func (m *State) Transit(newState string) {
	m.name = newState
	m.transited = true
}

func (m *State) RecentlyTransited() bool {
	return m.transited
}

func (m *State) Get(key string) string {
	return m.data[key]
}

func (m *State) Set(key string, value string) {
	m.data[key] = value
}
