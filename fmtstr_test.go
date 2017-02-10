package main

import (
	"testing"
)

func TestDefaultString(t *testing.T) {
	// since we don't actually do any parsing, this should suffice
	defstr := DefaultImporterSettings.String()
	if defstr != "f0000120001" {
		t.Fatal("importer string does not match expected")
	}
}
