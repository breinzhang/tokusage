package main

import (
	"os"

	"github.com/breinzhang/tokusage/internal/cli"
)

func main() {
	if err := cli.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
