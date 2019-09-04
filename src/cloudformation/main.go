package main

import (
	"aws-machete/src/cloudformation/cmd"
)

func main() {
	cmd.CommandManagerInstance.Execute()
}
