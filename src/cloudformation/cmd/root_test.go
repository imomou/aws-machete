package cmd

import (
	"testing"
)

func TestInitRootCmd(t *testing.T) {
	cm := initRootCmd()

	if cm == nil {
		t.Error("CommandManagement initialisation failed.")
	}
}
