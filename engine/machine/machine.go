package machine

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

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
	metrics                 *metrics.Command
	bus                     command.Bus
	stateDeterminer         *DialogDeterminer
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
		metrics:                 metrics.Command(),
		bus:                     bus,
		stateDeterminer:         newDeterminer(routes, stateStorage, metrics.Command()),
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
			h.metrics.IncNotFound()
			return h.commandNotFoundFallback(ctx, message)
		}

		return err
	}
	return nil
}

func (h *Machine) handle(ctx context.Context, message messengerapi.Message) error {
	slog.DebugContext(ctx, "[machine] handling message")

	dialog, err := h.stateDeterminer.Determine(ctx, message)
	if err != nil {
		return fmt.Errorf("determine command and state: %w", err)
	}

	startedAt := time.Now()

	ctx = logx.WithCommandName(ctx, dialog.Command.Name)

	slog.DebugContext(ctx, "[machine] find action")

	act, err := h.findAction(dialog.State, dialog.Command.Command)
	if err != nil {
		return fmt.Errorf("find action: %w", err)
	}

	slog.DebugContext(ctx, "[machine] action found", logx.StateName(act.State()))

	run := func(ctx context.Context, req *command.Request) error {
		if err = act.Run(ctx, req); err != nil {
			_, err = h.errorHandler(ctx, message, err)
			var codeErr command.CodeError
			if errors.As(err, &codeErr) {
				h.metrics.IncActionHandled(dialog.Command.Name, act.State(), codeErr.Code())
			}
		}

		return err
	}

	err = h.bus.Handle(ctx, &command.Request{
		Message: message,
		State:   dialog.State,
	}, run)
	if err != nil {
		return err
	}
	h.metrics.IncActionHandled(dialog.Command.Name, act.State(), "OK")

	if !dialog.State.RecentlyTransited() {
		nextAct := act.Next()
		// stop execution, because next action not found.
		if nextAct == nil {
			return h.finishState(ctx, act, dialog.State)
		}

		slog.InfoContext(
			ctx,
			"[machine] transit state",
			slog.String("from_state", act.State()),
			slog.String("next_state", nextAct.State()),
		)

		dialog.State.Transit(nextAct.State())
	}

	h.metrics.IncStateTransition(dialog.Command.Name, act.State(), dialog.State.Name())

	err = h.stateStorage.Put(ctx, dialog.State)
	if err != nil {
		return fmt.Errorf("put state: %w", err)
	}

	h.metrics.ObserveActionExecution(dialog.Command.Name, act.State(), time.Since(startedAt))

	if dialog.State.Forwarded() != nil {
		return h.forward(ctx, message, dialog.State, act)
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

	h.metrics.IncFinished(mState.CommandName())
	h.metrics.ObserveExecution(mState.CommandName(), mState.Duration())

	derr := h.stateStorage.Delete(ctx, mState)
	if derr != nil {
		return fmt.Errorf("delete state: %w", derr)
	}
	return nil
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
