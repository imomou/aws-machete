package cmd

import (
	"errors"
	"fmt"
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

	cfnManager := *uc.cm.cfnManager

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

	// Use template summary param to avoid missing out no echo parameters
	tempSummary, tempSummaryErr := cfnManager.getTemplateSummary(&templateString)
	if tempSummaryErr != nil {
		return tempSummaryErr
	}

	// Override parameters
	stackParams := make([]*cloudformation.Parameter, len(tempSummary.Parameters))
	for index, stackParam := range tempSummary.Parameters {
		parameterValue, exist := uc.params[*stackParam.ParameterKey]
		userPreviousValue := !exist
		if exist {
			stackParams[index] = &cloudformation.Parameter{
				ParameterKey:     stackParam.ParameterKey,
				ParameterValue:   &parameterValue,
				UsePreviousValue: &userPreviousValue,
			}
		} else {
			stackParams[index] = &cloudformation.Parameter{
				ParameterKey:     stackParam.ParameterKey,
				ParameterValue:   nil,
				UsePreviousValue: &userPreviousValue,
			}
		}
	}

	// Override tags
	stack, stackErr := cfnManager.getStack(&uc.target)
	if stackErr != nil {
		return stackErr
	}
	stackTags := make([]*cloudformation.Tag, len(stack.Tags))
	// existing tags
	for tagIndex, stackTag := range stack.Tags {
		if tagValue, exist := uc.tags[*stackTag.Key]; exist {
			stackTags[tagIndex] = &cloudformation.Tag{
				Key:   stackTag.Key,
				Value: &tagValue,
			}
		} else {
			stackTags[tagIndex] = &cloudformation.Tag{
				Key:   stackTag.Key,
				Value: stackTag.Value,
			}
		}
	}
	// new tags
	for ntKey, ntValue := range uc.tags {
		tagExist := func() bool {
			for _, stackTag := range stackTags {
				if ntKey == *stackTag.Key && ntValue == *stackTag.Value {
					return true
				}
			}
			return false
		}
		if !(tagExist()) {
			stackTags = append(stackTags, &cloudformation.Tag{Key: &ntKey, Value: &ntValue})
		}
	}

	// Create change set
	if uc.cm.config.mode == dry {
		// TODO print details.
		return nil
	}
	createCsOutput, createCsError := cfnManager.createChangeSet(&uc.target, stackParams, stackTags, &templateString, cloudformation.ChangeSetTypeUpdate)
	if createCsError != nil {
		return createCsError
	}

	if uc.cm.config.mode == changesetonly {
		return nil
	}

	// Execute change set
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
	executeChangeSetErr := cfnManager.executeChangeSet(createCsOutput.StackId, createCsOutput.Id)
	if executeChangeSetErr != nil {
		return executeChangeSetErr
	}
	return nil
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
