package msghandler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/artarts36/lowbot/pkg/engine/command"
	"github.com/artarts36/lowbot/pkg/engine/messenger"
	"github.com/artarts36/lowbot/pkg/engine/router"
	"github.com/artarts36/lowbot/pkg/engine/state"
)

type Handler struct {
	router       router.Router
	stateStorage state.Storage

	commandNotFoundFallback CommandNotFoundFallback
}

func NewHandler(
	routes router.Router,
	stateStorage state.Storage,
	commandNotFoundFallback CommandNotFoundFallback,
) *Handler {
	return &Handler{router: routes, stateStorage: stateStorage, commandNotFoundFallback: commandNotFoundFallback}
}

func (h *Handler) Handle(ctx context.Context, message messenger.Message) error {
	err := h.handle(ctx, message)
	if err != nil {
		if errors.Is(err, router.ErrCommandNotFound) {
			return h.commandNotFoundFallback(ctx, message)
		}

		return err
	}
	return nil
}

func (h *Handler) handle(ctx context.Context, message messenger.Message) error {
	slog.DebugContext(ctx, "[handler] handling message")

	cmd, mState, err := h.determineCommandAndState(ctx, message)
	if err != nil {
		return fmt.Errorf("determine command and state: %w", err)
	}

	slog.DebugContext(ctx, "[handler] find action", slog.String("command.name", cmd.Name))

	act, err := h.findAction(mState, cmd.Command)
	if err != nil {
		return fmt.Errorf("find action: %w", err)
	}

	slog.DebugContext(
		ctx,
		"[handler] action found",
		slog.String("command.name", cmd.Name),
		slog.String("state.name", act.State()),
	)

	err = act.Run(ctx, message, mState)
	if err != nil {
		validErr := &command.ValidationError{}
		if errors.As(err, &validErr) {
			return message.RespondText(validErr.Text)
		}

		return fmt.Errorf("run action: %w", err)
	}

	if !mState.RecentlyTransited() {
		nextAct := act.Next()
		if nextAct == nil {
			slog.InfoContext(
				ctx,
				"[handler] next state not found",
				slog.String("state.name", act.State()),
			)

			return h.stateStorage.Delete(ctx, mState)
		}

		slog.InfoContext(
			ctx,
			"[handler] transit state",
			slog.String("from_state", act.State()),
			slog.String("next_state", nextAct.State()),
		)

		mState.Transit(nextAct.State())
	}

	err = h.stateStorage.Put(ctx, mState)
	if err != nil {
		return fmt.Errorf("put state: %w", err)
	}

	return nil
}

func (h *Handler) determineCommandAndState(
	ctx context.Context,
	message messenger.Message,
) (*router.NamedCommand, *state.State, error) {
	var cmd *router.NamedCommand

	mState, err := h.stateStorage.Get(ctx, message.GetChatID())
	if err != nil {
		if errors.Is(err, state.ErrStateNotFound) {
			slog.DebugContext(ctx, "[handler] state not found")

			cmd, err = h.router.Find(message.ExtractCommandName())
			if err != nil {
				return nil, nil, err
			}

			slog.DebugContext(ctx, "[handler] command found", slog.String("command.name", cmd.Name))

			mState = state.NewState(message.GetChatID(), cmd.Name)

			return cmd, mState, nil
		} else {
			return nil, nil, fmt.Errorf("get state from storage: %w", err)
		}
	}

	slog.DebugContext(
		ctx,
		"[handler] found state",
		slog.String("state.name", mState.Name()),
		slog.String("state.command_name", mState.CommandName()),
	)

	cmd, err = h.router.Find(mState.CommandName())
	if err != nil {
		slog.ErrorContext(
			ctx,
			"[handler] failed to find command",
			slog.String("command.name", mState.CommandName()),
			slog.Any("err", err),
		)

		return nil, nil, fmt.Errorf("find command: %w", err)
	}

	slog.DebugContext(ctx, "[handler] command found", slog.String("command.name", cmd.Name))

	messageCommandName := message.ExtractCommandName()
	if messageCommandName != "" && messageCommandName != mState.CommandName() {
		allow, ierr := cmd.Command.Interrupt(ctx, message, mState.CommandName(), messageCommandName)
		if ierr != nil {
			return nil, nil, fmt.Errorf("defines interruption: %w", ierr)
		}

		if allow {
			cmd, err = h.router.Find(messageCommandName)
			if err != nil {
				return nil, nil, fmt.Errorf("find new command: %w", err)
			}

			slog.DebugContext(
				ctx,
				"[handler] interrupt command, switch to new command",
				slog.String("old_command.name", mState.CommandName()),
				slog.String("new_command.name", cmd.Name),
			)

			mState = state.NewState(message.GetChatID(), cmd.Name)
		}
	}

	return cmd, mState, nil
}

func (h *Handler) findAction(state *state.State, cmd command.Command) (command.Action, error) {
	acts := cmd.Actions()

	if state.Name() == "" {
		return acts.First(), nil
	}

	st, exists := acts.Get(state.Name())
	if !exists {
		return nil, fmt.Errorf("findAction not found")
	}

	return st, nil
}
