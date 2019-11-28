package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func (cm *CommandManagement) rootCmdRun(cmd *cobra.Command, args []string) {
	cmd.Help()
}

var rootCmdLong = `Route53 cli utility that can can do more sophisticated actions awscli cannot, such as copy a stack, etc.`

func initRootCmd() *CommandManagement {
	cm := &CommandManagement{
		config:         &config{},
		route53Manager: newRoute53client(),
		viper:          viper.New(),
	}
	cm.root = &cobra.Command{
		Use:   "route53",
		Short: "Route53 cli utility that does things awscli cannot.",
		Long:  rootCmdLong,
		Run:   cm.rootCmdRun,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

			fmt.Printf("Command: %#v\n", cmd.Name())
			cmd.Flags().VisitAll(func(flag *pflag.Flag) {
				cm.viper.BindPFlag(flag.Name, flag)
			})
			cm.viper.SetEnvPrefix("machete")
			cm.viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
			cm.viper.AutomaticEnv()

			configFile := cm.viper.GetString("config-path")
			if configFile != "" {
				// Use config file from the flag.
				fmt.Printf("Config file specified and found: %#v\n", configFile)
				configType := cm.viper.GetString("config-format")
				cm.viper.SetConfigFile(configFile)
				cm.viper.SetConfigType(configType)
				if vErr := cm.viper.ReadInConfig(); vErr != nil {
					return vErr
				}
			}

			config := cm.config
			config.timeout = cm.viper.GetInt("timeout")
			modeString := cm.viper.GetString("mode")
			fmt.Printf("Command execution mode: %v\n", modeString)
			config.mode = ParseMode(modeString)

			return nil
		},
	}
	// app flags, to be optionally overriden by viper.
	cm.root.PersistentFlags().StringP("mode", "m", "interactive", "Modes of command execution. Valid options are: noninteractive, changesetonly, dry, interactive.")
	cm.root.PersistentFlags().IntP("wait", "w", -1, "Time out in seconds to wait for the operation to complete. -1 means wait forever.")

	// viper flags.
	cm.root.PersistentFlags().StringP("config-path", "c", "", "Config file to supply flags / parameters with.")
	cm.root.PersistentFlags().String("config-format", "yaml", "Format of the configuration file.")

	cm.initGetAllCmd()
	cm.viper.SetKeysCaseSensitive(true)

	return cm
}

//CommandManagerInstance Init
var CommandManagerInstance CommandManager = initRootCmd()
