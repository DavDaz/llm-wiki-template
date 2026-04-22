package main

import (
	"os"

	"github.com/DavDaz/llm-wiki-template/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
