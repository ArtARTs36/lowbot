package router

import (
	"errors"

	"github.com/artarts36/lowbot/pkg/engine/command"
)

var (
	ErrCommandNotFound      = errors.New("command not found")
	ErrCommandAlreadyExists = errors.New("command already exists")
)

type Router interface {
	// Add command.
	// Throws ErrCommandAlreadyExists.
	Add(cmd *NamedCommand) error

	// Find command by message.
	// Throws ErrCommandNotFound.
	Find(cmdName string) (*NamedCommand, error)

	// List commands.
	List() []*NamedCommand
}

type NamedCommand struct {
	Name    string
	Command command.Command
}
