package msghandler

import (
	"context"
	"fmt"
	"github.com/agnivade/levenshtein"
	"github.com/artarts36/lowbot/pkg/engine/messenger"
	"github.com/artarts36/lowbot/pkg/engine/router"
	"strings"
)

type CommandNotFoundFallback func(ctx context.Context, message messenger.Message) error

func ErrorCommandNotFoundFallback() CommandNotFoundFallback {
	return func(ctx context.Context, message messenger.Message) error {
		return message.Respond(&messenger.Answer{
			Text: "Command not found.",
		})
	}
}

func SuggestCommandNotFoundFallback(routes router.Router) CommandNotFoundFallback {
	return func(ctx context.Context, message messenger.Message) error {
		msgCmd := message.ExtractCommandName()
		result := []string{
			fmt.Sprintf("Command \"%s\" not found.", msgCmd),
		}

		cmds := make([]string, 0)

		for _, cmd := range routes.List() {
			if levenshtein.ComputeDistance(msgCmd, cmd.Name) < 3 {
				cmds = append(cmds, fmt.Sprintf("/%s - %s", cmd.Name, cmd.Command.Description()))
			}
		}

		if len(cmds) > 0 {
			result = append(result, "")
			result = append(result, "Similar commands:")
			result = append(result, cmds...)
		}

		return message.Respond(&messenger.Answer{
			Text: strings.Join(result, "\n"),
		})
	}
}
