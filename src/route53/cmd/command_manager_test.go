package cmd

import (
	"testing"
)

func TestModeString(t *testing.T) {
	testModeValue := noninteractive

	modeString := testModeValue.String()

	if modeString != modes[0] {
		t.Error("String method failed.")
	}
}

func TestModeParse(t *testing.T) {
	testModeValue := modes[3]

	mode := ParseMode(testModeValue)

	if mode != 3 {
		t.Error("Parsing failed.")
	}
}
