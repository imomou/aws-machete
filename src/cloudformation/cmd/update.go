package cmd

import (
	"errors"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/spf13/cobra"
	"io/ioutil"
	"strings"
)

type updateCmd struct {
	target       string
	params       map[string]string
	tags         map[string]string
	templatePath string
	cm           *CommandManagement
	cmd          *cobra.Command
}

func (uc *updateCmd) runE(cmd *cobra.Command, args []string) error {

	cfnManager := uc.cm.cfnManager

	// Get template first.
	var templateString string
	if len(uc.templatePath) > 0 {
		buffer, templateReadErr := ioutil.ReadFile(uc.templatePath)
		if templateReadErr != nil {
			return templateReadErr
		} else {
			templateString = string(buffer)
		}
	} else {
		stackTemplate, stackTemplateErr := cfnManager.getStackTemplate(&uc.target)
		if stackTemplateErr != nil {
			return stackTemplateErr
		}
		templateString = *stackTemplate
	}

	// Parameters
	stackParams, spErr := uc.cm.filterParameters(&templateString, &uc.params)
	if spErr != nil {
		return spErr
	}

	// Override tags
	stack, stackErr := cfnManager.getStack(&uc.target)
	if stackErr != nil {
		return stackErr
	}

	stackTags := uc.cm.mergeTags(stack.Tags, &uc.tags)

	return uc.cm.createAndExecute(&uc.target, stackParams, stackTags, &templateString, cloudformation.ChangeSetTypeUpdate)
}

func (uc *updateCmd) preRunE(cmd *cobra.Command, args []string) error {

	localViper := uc.cm.viper
	uc.target = localViper.GetString("target")
	uc.params = localViper.GetStringMapString("param")
	uc.tags = localViper.GetStringMapString("tag")
	uc.templatePath = localViper.GetString("template-path")

	// parameter validations
	var errstrings []string
	if uc.target == "" {
		errstrings = append(errstrings, "Please specify target stack to update.")
	}
	if len(uc.params) == 0 && len(uc.tags) == 0 && uc.templatePath == "" {
		errstrings = append(errstrings, "Nothing specified to update.")
	}

	if len(errstrings) > 0 {
		return errors.New(strings.Join(errstrings, "\n"))
	}

	return nil
}

var updateCmdLong = `Perform a cloudformation update without having to respecify all the parameters and tags.`

func (cm *CommandManagement) initUpdateCmd() {

	// init command structure
	cmd := &cobra.Command{
		Use:   "update",
		Short: "update",
		Long:  updateCmdLong,
	}
	ucmd := &updateCmd{
		cm:  cm,
		cmd: cmd,
	}

	// local params
	cmd.Flags().StringP("target", "t", "", "Stack name or arn to update")
	cmd.Flags().StringToStringP("param", "p", nil, "Parameters to override")
	cmd.Flags().StringToStringP("tag", "g", nil, "Parameters to override")
	cmd.Flags().String("template-path", "", "Parameters to override")

	// wire methods.
	cmd.PreRunE = ucmd.preRunE
	cmd.RunE = ucmd.runE

	// register
	cm.root.AddCommand(cmd)
}
