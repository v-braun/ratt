package cmd

import (
	"errors"
	"os"

	"github.com/fatih/color"
)

func mustNoErr(err error) {
	if err != nil {
		color.New(color.Bold, color.FgGreen).Println("Execution failed with error:")
		color.New(color.Bold, color.FgGreen).Printf("%s\n", err)
		os.Exit(1)
	}
}
func fileExists(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		return true
	}
}
