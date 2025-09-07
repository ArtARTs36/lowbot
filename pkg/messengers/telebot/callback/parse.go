package callback

import "strings"

func ParseID(id string) ID {
	id = strings.TrimPrefix(id, "\f")

	parts := strings.SplitN(id, ":", 2)
	if len(parts) < 2 {
		return nil
	}

	switch parts[0] {
	case string(TypePassEnumValue):
		return NewPassEnumValue(parts[1])
	}

	return nil
}
