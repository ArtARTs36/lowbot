package command

type Definition struct {
	// This field user for command routing.
	Name string

	// This field may be used in /start command.
	Description string
}
