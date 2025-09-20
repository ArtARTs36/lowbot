package router

import (
	"errors"

	"github.com/artarts36/lowbot/engine/command"
)

var (
	ErrCommandNotFound      = errors.New("command not found")
	ErrCommandAlreadyExists = errors.New("command already exists")
)

type Router interface {
	// Add command.
	// Throws ErrCommandAlreadyExists.
	Add(cmd command.Command) error

	// Find command by message.
	// Throws ErrCommandNotFound.
	Find(cmdName string) (command.Command, error)

	// List commands.
	List() []command.Command
}
