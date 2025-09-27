package redisstatestorage

import (
	"time"

	"github.com/artarts36/lowbot/engine/state"
)

type stateRow struct {
	ChatID      string            `json:"chat_id"`
	Name        string            `json:"name"`
	CommandName string            `json:"command_name"`
	Data        map[string]string `json:"data"`
	StartedAt   time.Time         `json:"started_at"`
}

func (r *stateRow) from(st *state.State) {
	r.ChatID = st.ChatID()
	r.Name = st.Name()
	r.CommandName = st.CommandName()
	r.Data = st.All()
	r.StartedAt = st.StartedAt()
}

func (r *stateRow) state() *state.State {
	return state.NewFullState(r.ChatID, r.Name, r.CommandName, r.Data, r.StartedAt)
}
