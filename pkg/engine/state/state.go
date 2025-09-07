package state

import "time"

type State struct {
	chatID      string
	name        string
	commandName string

	transited bool

	data map[string]string

	startedAt time.Time
}

func NewState(chatID string, commandName string) *State {
	return &State{
		chatID:      chatID,
		commandName: commandName,
		data:        make(map[string]string),
		startedAt:   time.Now(),
	}
}

func NewFullState(
	chatID string,
	name string,
	commandName string,
	data map[string]string,
	startedAt time.Time,
) *State {
	return &State{
		chatID:      chatID,
		name:        name,
		commandName: commandName,
		data:        data,
		startedAt:   startedAt,
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

func (m *State) StartedAt() time.Time {
	return m.startedAt
}

func (m *State) Duration() time.Duration {
	return time.Since(m.startedAt)
}
