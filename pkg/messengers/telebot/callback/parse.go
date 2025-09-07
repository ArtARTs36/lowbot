package callback

import "strings"

const callbackPartsCount = 2

func ParseID(id string) ID {
	id = strings.TrimPrefix(id, "\f")

	parts := strings.SplitN(id, ":", callbackPartsCount)
	if len(parts) < callbackPartsCount {
		return nil
	}

	switch parts[0] { //nolint:gocritic // not need
	case string(TypePassEnumValue):
		return NewPassEnumValue(parts[1])
	}

	return nil
}
