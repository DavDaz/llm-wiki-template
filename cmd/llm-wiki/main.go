package main

import (
	"os"

	"github.com/DavDaz/llm-wiki-generator/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
