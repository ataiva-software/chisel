package main

import (
	"fmt"
	"os"

	"github.com/ataiva-software/chisel/pkg/cli"
)

var version = "dev"

func main() {
	if err := cli.Execute(version); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
