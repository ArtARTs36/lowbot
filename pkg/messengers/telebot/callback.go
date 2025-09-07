package telebot

import (
	"fmt"
	"strings"
)

type callbackType string

const (
	callbackTypeEnum callbackType = "enum"
)

type callbackID struct {
	ID string

	Type   callbackType
	Values []string
}

func parseCallbackID(id string) (callbackID, bool) {
	id = strings.TrimPrefix(id, "\f")

	parts := strings.SplitN(id, ":", 2)
	if len(parts) < 2 {
		return callbackID{}, false
	}

	return callbackID{
		ID:     id,
		Type:   callbackType(parts[0]),
		Values: parts[1:],
	}, true
}

func createEnumCallbackID(value string) callbackID {
	return callbackID{
		ID:     fmt.Sprintf("enum:%s", value),
		Type:   callbackTypeEnum,
		Values: []string{value},
	}
}
