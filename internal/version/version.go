package version

import "fmt"

const (
	Version string = "0.0.3"
)

func Print() {
	fmt.Printf("tsk v%s\n", Version)
}
