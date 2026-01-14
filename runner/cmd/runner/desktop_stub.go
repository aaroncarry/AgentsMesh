//go:build !desktop

package main

import (
	"fmt"
	"os"
)

// runDesktop is a stub when built without desktop tag.
// To enable desktop mode, build with: go build -tags desktop
func runDesktop(args []string) {
	fmt.Println("Desktop mode is not available in this build.")
	fmt.Println("To enable desktop mode, build with: go build -tags desktop")
	os.Exit(1)
}
