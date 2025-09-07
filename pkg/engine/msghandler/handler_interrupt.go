package msghandler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/artarts36/lowbot/pkg/engine/messenger"
	"github.com/artarts36/lowbot/pkg/engine/router"
	"github.com/artarts36/lowbot/pkg/engine/state"
)

func (h *Handler) detectInterrupt(message messenger.Message, mState *state.State) bool {
	messageCommandName := message.ExtractCommandName()

	return messageCommandName != "" && messageCommandName != mState.CommandName()
}

func (h *Handler) tryInterrupt(
	ctx context.Context,
	currentCommand *router.NamedCommand,
	message messenger.Message,
	mState *state.State,
) (*router.NamedCommand, *state.State, error) {
	desiredCommandName := message.ExtractCommandName()

	allow, err := currentCommand.Command.Interrupt(ctx, message, mState.CommandName(), desiredCommandName)
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
