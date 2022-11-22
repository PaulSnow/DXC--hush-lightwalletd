// Copyright 2021 The Hush developers
// Released under the GPLv3
package main

import (
	"os"
	"testing"
)

// TestFileExists checks whether or not the file exists
func TestFileExists(t *testing.T) {
	if fileExists("nonexistent-file") {
		t.Fatal("fileExists unexpected success")
	}
	// If the path exists but is a directory, should return false
	if fileExists(".") {
		t.Fatal("fileExists unexpected success")
	}
	// The following file should exist, it's what's being tested
	if !fileExists("main.go") {
		t.Fatal("fileExists failed")
	}
}

// fileExists checks if file exists and is not directory to prevent further errors
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
