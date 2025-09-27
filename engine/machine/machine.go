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
	logger                  logx.Logger
}

func New(
	routes router.Router,
	stateStorage state.Storage,
	errorHandler ErrorHandler,
	commandNotFoundFallback CommandNotFoundFallback,
	metrics *metrics.Group,
	bus command.Bus,
	logger logx.Logger,
) *Machine {
	return &Machine{
		router:                  routes,
		stateStorage:            stateStorage,
		errorHandler:            errorHandler,
		commandNotFoundFallback: commandNotFoundFallback,
		metrics:                 metrics.Command(),
		bus:                     bus,
		stateDeterminer:         newDeterminer(routes, stateStorage, metrics.Command(), logger),
		logger:                  logger,
	}
}

type Request struct {
	Message   messengerapi.Message
	Responder messengerapi.Responder
}

func (h *Machine) Handle(ctx context.Context, req *Request) error {
	ctx = logx.WithMessageID(
		logx.WithChatID(ctx, req.Message.GetChatID()),
		req.Message.GetID(),
	)

	err := h.handle(ctx, req)
	if err != nil {
		if errors.Is(err, router.ErrCommandNotFound) {
			h.metrics.IncNotFound()
			return h.commandNotFoundFallback(ctx, req)
		}

		return err
	}
	return nil
}

func (h *Machine) handle(ctx context.Context, req *Request) error {
	h.logger.DebugContext(ctx, "[machine] handling message")

	dialog, err := h.stateDeterminer.Determine(ctx, req.Message)
	if err != nil {
		return fmt.Errorf("determine command and state: %w", err)
	}

	startedAt := time.Now()

	ctx = logx.WithCommandName(ctx, dialog.Command.Definition().Name)

	h.logger.DebugContext(ctx, "[machine] find action")

	act, err := h.findAction(dialog.State, dialog.Command)
	if err != nil {
		return fmt.Errorf("find action: %w", err)
	}

	h.logger.DebugContext(ctx, "[machine] action found", logx.StateName(act.State()))

	run := func(ctx context.Context, cmdReq *command.Request) error {
		if err = act.Run(ctx, cmdReq); err != nil {
			_, err = h.errorHandler(ctx, req, err)
			var codeErr command.CodeError
			if errors.As(err, &codeErr) {
				h.metrics.IncActionHandled(dialog.Command.Definition().Name, act.State(), codeErr.Code())

				return fmt.Errorf("%s: %w", codeErr.Code(), codeErr)
			}
		}

		return err
	}

	err = h.bus.Handle(ctx, &command.Request{
		Message:   req.Message,
		Responder: req.Responder,
		State:     dialog.State,
	}, run)
	if err != nil {
		return err
	}
	h.metrics.IncActionHandled(dialog.Command.Definition().Name, act.State(), "OK")

	if !dialog.State.RecentlyTransited() {
		nextAct := act.Next()
		// stop execution, because next action not found.
		if nextAct == nil {
			return h.finishState(ctx, act, dialog.State)
		}

		h.logger.InfoContext(
			ctx,
			"[machine] transit state",
			slog.String("from_state", act.State()),
			slog.String("next_state", nextAct.State()),
		)

		dialog.State.Transit(nextAct.State())
	}

	h.metrics.IncStateTransition(dialog.Command.Definition().Name, act.State(), dialog.State.Name())

	err = h.stateStorage.Put(ctx, dialog.State)
	if err != nil {
		return fmt.Errorf("put state: %w", err)
	}

	h.metrics.ObserveActionExecution(dialog.Command.Definition().Name, act.State(), time.Since(startedAt))

	if dialog.State.Forwarded() != nil {
		return h.forward(ctx, req, dialog.State, act)
	}

	return nil
}

func (h *Machine) forward(
	ctx context.Context,
	req *Request,
	mState *state.State,
	act command.Action,
) error {
	if mState.Forwarded().NewStateName() == state.Passthrough {
		h.logger.DebugContext(ctx, "[lowbot][machine] passthrough", slog.String("to_state", act.Next().State()))
	} else {
		h.logger.DebugContext(ctx, "[lowbot][machine] forwarding", slog.String("to_state", act.State()))
	}

	return h.handle(ctx, req)
}

func (h *Machine) finishState(ctx context.Context, act command.Action, mState *state.State) error {
	h.logger.InfoContext(ctx, "[lowbot][machine] next state not found", slog.String("state.name", act.State()))

	h.metrics.IncFinished(mState.CommandName())
	h.metrics.ObserveExecution(mState.CommandName(), mState.Duration())

	err := h.stateStorage.Delete(ctx, mState)
	if err != nil {
		if errors.Is(err, state.ErrStateNotFound) {
			return nil
		}

		return fmt.Errorf("delete state: %w", err)
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
