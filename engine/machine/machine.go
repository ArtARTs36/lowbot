package machine

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/artarts36/lowbot/metrics"

	"github.com/artarts36/lowbot/logx"

	"github.com/artarts36/lowbot/engine/command"
	"github.com/artarts36/lowbot/engine/router"
	"github.com/artarts36/lowbot/engine/state"
	"github.com/artarts36/lowbot/messenger/messengerapi"
)

type Machine struct {
	router       router.Router
	stateStorage state.Storage
	errorHandler ErrorHandler

	commandNotFoundFallback CommandNotFoundFallback
	metrics                 *metrics.Group
	bus                     command.Bus
}

func New(
	routes router.Router,
	stateStorage state.Storage,
	errorHandler ErrorHandler,
	commandNotFoundFallback CommandNotFoundFallback,
	metrics *metrics.Group,
	bus command.Bus,
) *Machine {
	return &Machine{
		router:                  routes,
		stateStorage:            stateStorage,
		errorHandler:            errorHandler,
		commandNotFoundFallback: commandNotFoundFallback,
		metrics:                 metrics,
		bus:                     bus,
	}
}

func (h *Machine) Handle(ctx context.Context, message messengerapi.Message) error {
	ctx = logx.WithMessageID(
		logx.WithChatID(ctx, message.GetChatID()),
		message.GetID(),
	)

	err := h.handle(ctx, message)
	if err != nil {
		if errors.Is(err, router.ErrCommandNotFound) {
			h.metrics.IncCommandNotFound()
			return h.commandNotFoundFallback(ctx, message)
		}

		return err
	}
	return nil
}

func (h *Machine) handle(ctx context.Context, message messengerapi.Message) error {
	slog.DebugContext(ctx, "[machine] handling message")

	cmd, mState, err := h.determineCommandAndState(ctx, message)
	if err != nil {
		return fmt.Errorf("determine command and state: %w", err)
	}

	ctx = logx.WithCommandName(ctx, cmd.Name)

	slog.DebugContext(ctx, "[machine] find action")

	act, err := h.findAction(mState, cmd.Command)
	if err != nil {
		return fmt.Errorf("find action: %w", err)
	}

	slog.DebugContext(ctx, "[machine] action found", logx.StateName(act.State()))

	err = h.bus.Handle(ctx, &command.Request{
		Message: message,
		State:   mState,
	}, act)
	if err != nil {
		_, err = h.errorHandler(ctx, message, err)
		return err
	}

	if !mState.RecentlyTransited() {
		nextAct := act.Next()
		// stop execution, because next action not found.
		if nextAct == nil {
			return h.finishState(ctx, act, mState)
		}

		slog.InfoContext(
			ctx,
			"[machine] transit state",
			slog.String("from_state", act.State()),
			slog.String("next_state", nextAct.State()),
		)

		mState.Transit(nextAct.State())
	}

	h.metrics.IncCommandStateTransition(cmd.Name, act.State(), mState.Name())

	err = h.stateStorage.Put(ctx, mState)
	if err != nil {
		return fmt.Errorf("put state: %w", err)
	}

	if mState.Forwarded() != nil {
		return h.forward(ctx, message, mState, act)
	}

	return nil
}

func (h *Machine) forward(
	ctx context.Context,
	message messengerapi.Message,
	mState *state.State,
	act command.Action,
) error {
	if mState.Forwarded().NewStateName() == state.Passthrough {
		slog.DebugContext(ctx, "[machine] passthrough", slog.String("to_state", act.Next().State()))
	} else {
		slog.DebugContext(ctx, "[machine] forwarding", slog.String("to_state", act.State()))
	}

	return h.handle(ctx, message)
}

func (h *Machine) finishState(ctx context.Context, act command.Action, mState *state.State) error {
	slog.InfoContext(ctx, "[machine] next state not found", slog.String("state.name", act.State()))

	h.metrics.IncCommandFinished(mState.CommandName())
	h.metrics.ObserveCommandExecution(mState.CommandName(), mState.Duration())

	derr := h.stateStorage.Delete(ctx, mState)
	if derr != nil {
		return fmt.Errorf("delete state: %w", derr)
	}
	return nil
}

func (h *Machine) determineCommandAndState(
	ctx context.Context,
	message messengerapi.Message,
) (*router.NamedCommand, *state.State, error) {
	var cmd *router.NamedCommand

	mState, err := h.stateStorage.Get(ctx, message.GetChatID())
	if err != nil {
		if errors.Is(err, state.ErrStateNotFound) {
			cmdName := message.ExtractCommandName()

			slog.DebugContext(ctx, "[machine] state not found", slog.String("command.name", cmdName))

			cmd, err = h.router.Find(cmdName)
			if err != nil {
				return nil, nil, err
			}

			slog.DebugContext(ctx, "[machine] command found", slog.String("command.name", cmd.Name))

			mState = state.NewState(message.GetChatID(), cmd.Name)

			return cmd, mState, nil
		}

		return nil, nil, fmt.Errorf("get state from storage: %w", err)
	}

	slog.DebugContext(ctx,
		"[machine] found state",
		logx.StateName(mState.Name()),
		slog.String("state.command_name", mState.CommandName()),
	)

	cmd, err = h.router.Find(mState.CommandName())
	if err != nil {
		slog.ErrorContext(ctx, "[machine] failed to find command", logx.CommandName(mState.CommandName()), logx.Err(err))

		return nil, nil, fmt.Errorf("find command: %w", err)
	}

	slog.DebugContext(ctx, "[machine] command found", logx.CommandName(cmd.Name))

	if h.detectInterrupt(message, mState) {
		slog.DebugContext(ctx, "[machine] interrupt detected", logx.CommandName(cmd.Name))

		cmd, mState, err = h.tryInterrupt(ctx, cmd, message, mState)
		if err != nil {
			return nil, nil, fmt.Errorf("try interrupt: %w", err)
		}
	}

	return cmd, mState, nil
}

func (h *Machine) findAction(state *state.State, cmd command.Command) (command.Action, error) {
	acts := cmd.Actions()

	if state.Name() == "" {
		return acts.First(), nil
	}

	st, exists := acts.Get(state.Name())
	if !exists {
		return nil, fmt.Errorf("action not found")
	}

	return st, nil
}
