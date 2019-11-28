package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type CommandManager interface {
	Execute() error
}

type CommandManagement struct {
	root           *cobra.Command
	config         *config
	route53Manager route53Management
	viper          *viper.Viper
}

type config struct {
	mode    mode
	timeout int
}

type mode int

const (
	noninteractive mode = iota
	changesetonly
	dry
	interactive
)

var modes = [...]string{
	"noninteractive",
	"changesetonly",
	"dry",
	"interactive",
}

func (m mode) String() string {
	return modes[m]
}
func ParseMode(value string) mode {
	for r, m := range modes {
		if m == value {
			return mode(r)
		}
	}
	return -1
}

func (cm *CommandManagement) Execute() error {
	return cm.root.Execute()
}
