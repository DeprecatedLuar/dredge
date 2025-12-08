package main

import (
	"os"

	"github.com/DeprecatedLuar/dredge/internal/commands/parser"
)

func main() {
	parser.Route(os.Args[1:])
}
