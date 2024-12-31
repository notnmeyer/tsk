package outputformat

import (
	"fmt"
)

type OutputFormat string

const (
	JSON     OutputFormat = "json"
	Markdown OutputFormat = "md"
	Text     OutputFormat = "text"
	TOML     OutputFormat = "toml"
)

func String() string {
	return fmt.Sprintf("%s, %s, %s, %s", string(JSON), string(Markdown), string(Text), string(TOML))
}

func IsValid(format string) bool {
	switch format {
	case string(JSON), string(Markdown), string(Text), string(TOML):
		return true
	}
	return false
}
