package cmd

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/spf13/cobra"
)

type deleteAllCmd struct {
	cm  *CommandManagement
	cmd *cobra.Command
}

func (uc *deleteAllCmd) runE(cmd *cobra.Command, args []string) error {

	cfnManager := *uc.cm.cfnManager

	stackChannel := make(chan *cloudformation.Stack)
	defer close(stackChannel)
	errChannel := make(chan error)
	defer close(errChannel)

	go func() {
		cfnManager.getAll(stackChannel, errChannel)
	}()

	stacks := make([]*cloudformation.Stack, 0)
	for i := 0; i < cfnManager.getRegionCount(); {
		select {
		case err := <-errChannel:
			return err
		case stack := <-stackChannel:
			if stack == nil {
				i = i + 1
				continue
			}

			if *stack.EnableTerminationProtection {
				continue
			}

			stacks = append(stacks, stack)

			fmt.Printf("Deleting stack: %v (%v - %v)\n", *stack.StackName, getRegionFromArn(stack.StackId), *stack.StackStatus)
			if uc.cm.config.mode == interactive {
				fmt.Print("Please type \"confirm\" to proceed...")
				var confirmString string
				fmt.Scanf("%s", &confirmString)
				if confirmString == "confirm" {
					fmt.Println("Confirmed. Command resuming...")
				} else {
					fmt.Println("Confirmation failed. Exiting...")
					return nil
				}
			}

			if uc.cm.config.mode == dry {
				continue
			}

			delErr := cfnManager.delete(stack.StackId)
			if delErr != nil {
				return delErr
			}
		}
	}

	if uc.cm.config.mode == dry {
		fmt.Println("This is a dry run. No stacks were deleted.")
	}

	return nil
}

func (uc *deleteAllCmd) preRunE(cmd *cobra.Command, args []string) error {

	if uc.cm.config.mode == changesetonly {
		return errors.New("Mode changesetonly is not allowed for delete-all cmd.")
	}

	return nil
}

var deleteAllCmdLong = `Delete all stacks in all regions.`

func (cm *CommandManagement) initDeleteAllCmd() {

	// init command structure
	cmd := &cobra.Command{
		Use:   "delete-all",
		Short: "delete-all",
		Long:  deleteAllCmdLong,
	}
	cmdContainer := &deleteAllCmd{
		cm:  cm,
		cmd: cmd,
	}

	// wire methods.
	cmd.PreRunE = cmdContainer.preRunE
	cmd.RunE = cmdContainer.runE

	// register
	cm.root.AddCommand(cmd)
}
