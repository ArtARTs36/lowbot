package machine

import (
	"context"
	"fmt"
	"strings"

	"github.com/artarts36/lowbot/engine/router"
	"github.com/artarts36/lowbot/messenger/messengerapi"

	"github.com/agnivade/levenshtein"
)

const levenshteinThreshold = 3

type CommandNotFoundFallback func(ctx context.Context, req *Request) error

func ErrorCommandNotFoundFallback() CommandNotFoundFallback {
	return func(_ context.Context, req *Request) error {
		_, err := req.Responder.Respond(&messengerapi.Answer{
			Text: "Command not found.",
		})
		return err
	}
}

func SuggestCommandNotFoundFallback(routes router.Router) CommandNotFoundFallback {
	return func(_ context.Context, req *Request) error {
		msgCmd := req.Message.ExtractCommandName()
		result := []string{
			fmt.Sprintf("Command \"%s\" not found.", msgCmd),
		}

		cmds := make([]string, 0)

		for _, cmd := range routes.List() {
			if levenshtein.ComputeDistance(msgCmd, cmd.Definition().Name) < levenshteinThreshold {
				cmds = append(cmds, fmt.Sprintf("/%s - %s", cmd.Definition().Name, cmd.Definition().Description))
			}
		}

		if len(cmds) > 0 {
			result = append(result, "")
			result = append(result, "Similar commands:")
			result = append(result, cmds...)
		}

		_, err := req.Responder.Respond(&messengerapi.Answer{
			Text: strings.Join(result, "\n"),
		})
		return err
	}
}
