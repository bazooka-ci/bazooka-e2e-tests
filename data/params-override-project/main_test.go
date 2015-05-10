package main

import (
	"os"

	"testing"
)

func TestParse(t *testing.T) {
	if answer := os.Getenv("PARAM"); answer != "42" {
		t.Fatalf("Error: wanted '42', got '%s", answer)
	}
}
