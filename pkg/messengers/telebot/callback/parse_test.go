package callback

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestParse(t *testing.T) {
	cases := []struct {
		Title    string
		ID       string
		Expected ID
	}{
		{
			Title: "PassEnumValue",
			ID:    "PassEnumValue:int",
			Expected: &PassEnumValue{
				Value: "int",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Title, func(t *testing.T) {
			assert.Equal(t, c.Expected, ParseID(c.ID))
		})
	}
}
