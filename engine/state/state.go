package state

import "time"

const (
	Passthrough = "lowbot.passthrough" //nolint:gosec //false-positive
)

type State struct {
	chatID      string
	name        string
	commandName string

	forward   *Forward
	transited bool

	data map[string]string

	startedAt time.Time
}

type Forward struct {
	newStateName string
}

func (f *Forward) NewStateName() string {
	return f.newStateName
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
	if data == nil {
		data = make(map[string]string)
	}

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

// Forward save current state and immediately run action for state {newStateName}.
func (m *State) Forward(newStateName string) {
	m.forward = &Forward{
		newStateName: newStateName,
	}
	m.Transit(newStateName)
}

// Passthrough save current state and immediately run action for next state.
func (m *State) Passthrough() error {
	m.forward = &Forward{
		newStateName: Passthrough,
	}

	return nil
}

func (m *State) Forwarded() *Forward {
	return m.forward
}
