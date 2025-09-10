package machine

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/artarts36/lowbot/engine/command"

	"github.com/artarts36/lowbot/engine/router"
	"github.com/artarts36/lowbot/engine/state"
	"github.com/artarts36/lowbot/messenger/messengerapi"
)

func (h *Machine) detectInterrupt(message messengerapi.Message, mState *state.State) bool {
	messageCommandName := message.ExtractCommandName()

	return messageCommandName != "" && messageCommandName != mState.CommandName()
}

func (h *Machine) tryInterrupt(
	ctx context.Context,
	currentCommand *router.NamedCommand,
	message messengerapi.Message,
	mState *state.State,
) (*router.NamedCommand, *state.State, error) {
	desiredCommandName := message.ExtractCommandName()

	allow, err := currentCommand.Command.Interrupt(ctx, &command.InterruptRequest{
		Message:      message,
		CurrentState: mState.CommandName(),
		NewCommand:   desiredCommandName,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("defines interruption: %w", err)
	}

	h.metrics.IncCommandInterruption(currentCommand.Name, mState.Name(), desiredCommandName, allow)

	if !allow {
		return currentCommand, mState, nil
	}

	newCommand, err := h.router.Find(desiredCommandName)
	if err != nil {
		return nil, nil, fmt.Errorf("find new command: %w", err)
	}

	slog.DebugContext(ctx,
		"[handler] interrupt command, switch to new command",
		slog.String("from_command.name", currentCommand.Name),
		slog.String("to_command.name", newCommand.Name),
	)

	newState := state.NewState(message.GetChatID(), newCommand.Name)

	return newCommand, newState, nil
}
