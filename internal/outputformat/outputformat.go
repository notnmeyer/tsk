package outputformat

import (
	"fmt"
)

type OutputFormat string

const (
	Markdown OutputFormat = "md"
	Text     OutputFormat = "text"
	TOML     OutputFormat = "toml"
)

func String() string {
	return fmt.Sprintf("%s, %s, %s", string(Markdown), string(Text), string(TOML))
}

func IsValid(format string) bool {
	switch format {
	case string(Markdown), string(Text), string(TOML):
		return true
	}
	return false
}
