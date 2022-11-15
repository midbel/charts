package decode

import (
	"fmt"
)

type OptionError struct {
	Option  string
	Section string
	File    string
	Position
}

func (e OptionError) Error() string {
	return fmt.Sprintf("%s[%s]: option %s not recognized in section %s", e.File, e.Position, e.Option, e.Section)
}

type DecodeError struct {
	Message string
	File    string
	Position
}

func (e DecodeError) Error() string {
	return fmt.Sprintf("%s[%s]: %s", e.File, e.Position, e.Message)
}
