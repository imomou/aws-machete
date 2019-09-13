package cmd

import (
	"errors"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/spf13/cobra"
	"io/ioutil"
	"strings"
)

type ensureCmd struct {
	target       string
	params       map[string]string
	tags         map[string]string
	templatePath string
	cm           *CommandManagement
	cmd          *cobra.Command
}

func (uc *ensureCmd) runE(cmd *cobra.Command, args []string) error {

	cfnManager := uc.cm.cfnManager

	stack, stackErr := cfnManager.getStack(&uc.target)
	if stackErr != nil {
		return stackErr
	}

	// determine if the stack exists.
	csType := cloudformation.ChangeSetTypeUpdate
	if stack == nil {
		// stack not found. Create.
		csType = cloudformation.ChangeSetTypeCreate
	}

	// Get template first.
	var templateString string
	if len(uc.templatePath) > 0 {
		// template specified.
		buffer, templateReadErr := ioutil.ReadFile(uc.templatePath)
		if templateReadErr != nil {
			return templateReadErr
		} else {
			templateString = string(buffer)
		}
	} else if stack != nil {
		// template not specified
		// stack found
		stackTemplate, stackTemplateErr := cfnManager.getStackTemplate(&uc.target)
		if stackTemplateErr != nil {
			return stackTemplateErr
		}
		templateString = *stackTemplate
	} else {
		//template not specified
		// stack not found

		return errors.New("No cloudformation template specified and no stack found. Cannot proceed.")
	}

	// Parameters
	stackParams, spErr := uc.cm.filterParameters(&templateString, &uc.params)
	if spErr != nil {
		return spErr
	}

	// Override tags
	oldTags := make([]*cloudformation.Tag, 0)
	if stack != nil {
		oldTags = stack.Tags
	}
	stackTags := uc.cm.mergeTags(oldTags, &uc.tags)

	return uc.cm.createAndExecute(&uc.target, stackParams, stackTags, &templateString, csType)
}

func (uc *ensureCmd) preRunE(cmd *cobra.Command, args []string) error {

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

var ensureCmdLong = `Create a cloudformation stack if the specified stack does not exist. Otherwise, perform an update.`

func (cm *CommandManagement) initEnsureCmd() {

	// init command structure
	cmd := &cobra.Command{
		Use:   "ensure",
		Short: "ensure",
		Long:  ensureCmdLong,
	}
	ucmd := &ensureCmd{
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
