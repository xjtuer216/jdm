package main

import (
	"github.com/xjtuer216/jdm/cmd"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
