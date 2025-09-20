package machine

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/artarts36/lowbot/engine/command"
	"github.com/artarts36/lowbot/engine/router"
	"github.com/artarts36/lowbot/engine/state"
	"github.com/artarts36/lowbot/logx"
	"github.com/artarts36/lowbot/messenger/messengerapi"
	"github.com/artarts36/lowbot/metrics"
)

type Dialog struct {
	State   *state.State
	Command *router.NamedCommand
}

type DialogDeterminer struct {
	router       router.Router
	stateStorage state.Storage
	metrics      *metrics.Command
	steps        []determineStep
	logger       logx.Logger
}

type determineStep struct {
	name string

	fn func(ctx context.Context, message messengerapi.Message, env *determineStepEnv) (stop bool, err error)
}

type determineStepEnv struct {
	command *router.NamedCommand
	state   *state.State
}

func newDeterminer(
	router router.Router,
	stateStorage state.Storage,
	metrics *metrics.Command,
	logger logx.Logger,
) *DialogDeterminer {
	dd := &DialogDeterminer{
		router:       router,
		stateStorage: stateStorage,
		metrics:      metrics,
		steps:        []determineStep{},
		logger:       logger,
	}

	dd.steps = []determineStep{
		{
			name: "get state and command from message arguments (button calls, etc.)",
			fn:   dd.determineFromMessageArgs,
		},
		{
			name: "get state from storage",
			fn:   dd.getStateFromStorage,
		},
		{
			name: "if new dialog create new state",
			fn:   dd.tryCreateNewDialog,
		},
		{
			name: "continue dialog with found state",
			fn:   dd.continueDialog,
		},
	}

	return dd
}

func (h *DialogDeterminer) Determine(
	ctx context.Context,
	message messengerapi.Message,
) (*Dialog, error) {
	h.logger.DebugContext(ctx, "[lowbot][machine][determiner] determining dialog")

	env := &determineStepEnv{}
	for _, step := range h.steps {
		h.logger.DebugContext(ctx, "[lowbot][machine][determiner] running step", slog.String("step.name", step.name))

		stop, err := step.fn(ctx, message, env)
		if err != nil {
			return nil, err
		}
		if stop {
			return &Dialog{
				State:   env.state,
				Command: env.command,
			}, nil
		}
	}

	return nil, errors.New("dialog not determined")
}

func (h *DialogDeterminer) determineFromMessageArgs(
	ctx context.Context,
	message messengerapi.Message,
	env *determineStepEnv,
) (bool, error) {
	args := message.GetArgs()
	if args == nil {
		return false, nil
	}

	h.logger.DebugContext(ctx, "[lowbot][machine] message has predefined state", slog.Any("args", message.GetArgs()))

	mState := state.NewFullState(message.GetChatID(), args.StateName, args.CommandName, args.Data, time.Now())
	cmd, err := h.router.Find(args.CommandName)
	if err != nil {
		return true, err
	}

	env.state = mState
	env.command = cmd

	return true, nil
}

func (h *DialogDeterminer) getStateFromStorage(
	ctx context.Context,
	message messengerapi.Message,
	env *determineStepEnv,
) (bool, error) {
	dState, err := h.stateStorage.Get(ctx, message.GetChatID())
	if err != nil && !errors.Is(err, state.ErrStateNotFound) {
		return true, err
	}

	if dState != nil {
		h.logger.DebugContext(ctx,
			"[machine] found state",
			logx.StateName(dState.Name()),
			slog.String("state.command_name", dState.CommandName()),
		)
		env.state = dState
	}

	return false, nil
}

func (h *DialogDeterminer) tryCreateNewDialog(
	ctx context.Context,
	message messengerapi.Message,
	env *determineStepEnv,
) (bool, error) {
	if env.state != nil {
		return false, nil
	}

	cmdName := message.ExtractCommandName()

	h.logger.DebugContext(ctx, "[machine] state not found", slog.String("command.name", cmdName))

	cmd, err := h.router.Find(cmdName)
	if err != nil {
		return true, err
	}

	h.logger.DebugContext(ctx, "[machine] command found", slog.String("command.name", cmd.Name))

	env.command = cmd
	env.state = state.NewState(message.GetChatID(), cmd.Name)

	return true, nil
}

func (h *DialogDeterminer) continueDialog(
	ctx context.Context,
	message messengerapi.Message,
	env *determineStepEnv,
) (bool, error) {
	cmd, err := h.router.Find(env.state.CommandName())
	if err != nil {
		h.logger.ErrorContext(ctx, "[machine] failed to find command",
			logx.CommandName(env.state.CommandName()),
			logx.Err(err),
		)

		return true, fmt.Errorf("find command: %w", err)
	}

	h.logger.DebugContext(ctx, "[machine] command found", logx.CommandName(cmd.Name))

	if h.detectInterrupt(message, env.state) {
		h.logger.DebugContext(ctx, "[machine] interrupt detected", logx.CommandName(cmd.Name))

		cmd, env.state, err = h.tryInterrupt(ctx, cmd, message, env.state)
		if err != nil {
			return true, fmt.Errorf("try interrupt: %w", err)
		}
	}

	env.command = cmd

	return true, nil
}

func (h *DialogDeterminer) detectInterrupt(message messengerapi.Message, mState *state.State) bool {
	messageCommandName := message.ExtractCommandName()

	return messageCommandName != "" && messageCommandName != mState.CommandName()
}

func (h *DialogDeterminer) tryInterrupt(
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

	h.metrics.IncInterruption(currentCommand.Name, mState.Name(), desiredCommandName, allow)

	if !allow {
		return currentCommand, mState, nil
	}

	newCommand, err := h.router.Find(desiredCommandName)
	if err != nil {
		return nil, nil, fmt.Errorf("find new command: %w", err)
	}

	h.logger.DebugContext(ctx,
		"[handler] interrupt command, switch to new command",
		slog.String("from_command.name", currentCommand.Name),
		slog.String("to_command.name", newCommand.Name),
	)

	newState := state.NewState(message.GetChatID(), newCommand.Name)

	return newCommand, newState, nil
}
