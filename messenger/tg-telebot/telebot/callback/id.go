package callback

type Type string

const (
	TypePassEnumValue Type = "PassEnumValue"
)

type ID interface {
	String() string
}
