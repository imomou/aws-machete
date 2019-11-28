package cmd

import (
	// "errors"
	"fmt"
	//"github.com/aws/aws-sdk-go/service/route53"
	"github.com/spf13/cobra"
)

type getAllCmd struct {
	cm  *CommandManagement
	cmd *cobra.Command
}

func (uc *getAllCmd) runE(cmd *cobra.Command, args []string) error {

	route53Manager := uc.cm.route53Manager

	results, err := route53Manager.GetAllRecordSets()

	if err != nil {
		return err
	}

	for _, result := range results {
		fmt.Printf("%s: %10s \n", *result.Type, *result.Name)
	}

	return nil
}

// do I nee this?
func (uc *getAllCmd) preRunE(cmd *cobra.Command, args []string) error {
	return nil
}

var getAllCmdLong = `Get all host record sets in all regions`

func (cm *CommandManagement) initGetAllCmd() {

	// init command structure
	cmd := &cobra.Command{
		Use:   "get-all",
		Short: "get-all",
		Long:  getAllCmdLong,
	}

	cmdContainer := &getAllCmd{
		cm:  cm,
		cmd: cmd,
	}

	// wire methods
	cmd.PreRunE = cmdContainer.preRunE
	cmd.RunE = cmdContainer.runE

	// register
	cm.root.AddCommand(cmd)
}
