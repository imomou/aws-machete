package main

import (
	"aws-machete/src/route53/cmd"
)

func main() {
	cmd.CommandManagerInstance.Execute()
}
