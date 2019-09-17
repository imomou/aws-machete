package cmd

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	//"strings"
)

func (cm *CommandManagement) createAndExecute(
	stackName *string, params []*cloudformation.Parameter, tags []*cloudformation.Tag, templateBody *string, changeSetType string) error {

	// Create change set
	if cm.config.mode == dry {
		// TODO print details.
		return nil
	}
	createCsOutput, createCsError := cm.cfnManager.createChangeSet(stackName, params, tags, templateBody, changeSetType)
	if createCsError != nil {
		return createCsError
	}

	if cm.config.mode == changesetonly {
		return nil
	}

	// Execute change set
	if cm.config.mode == interactive {
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
	return cm.cfnManager.executeChangeSet(createCsOutput.StackId, createCsOutput.Id)
}

func (cm *CommandManagement) filterParameters(templateBody *string, values *map[string]string, isUpdate bool) ([]*cloudformation.Parameter, error) {
	tempSummary, tempSummaryErr := cm.cfnManager.getTemplateSummary(templateBody)
	if tempSummaryErr != nil {
		return nil, tempSummaryErr
	}

	stackParams := make([]*cloudformation.Parameter, len(tempSummary.Parameters))
	fmt.Printf("Keys: %#v\n", values)
	for index, stackParam := range tempSummary.Parameters {
		parameterValue, exist := (*values)[*stackParam.ParameterKey]
		//userPreviousValue := !exist
		if exist {
			fmt.Printf("Param key: %#v\n", stackParam.ParameterKey)
			stackParams[index] = &cloudformation.Parameter{
				ParameterKey:     stackParam.ParameterKey,
				ParameterValue:   &parameterValue,
				UsePreviousValue: aws.Bool(false),
			}
		} else {
			stackParams[index] = &cloudformation.Parameter{
				ParameterKey:   stackParam.ParameterKey,
				ParameterValue: nil,
			}
			if isUpdate {
				stackParams[index].UsePreviousValue = aws.Bool(true)
			}
		}
	}

	return stackParams, nil
}

func (cm *CommandManagement) mergeTags(oldTags []*cloudformation.Tag, newTags *map[string]string) []*cloudformation.Tag {
	stackTags := make([]*cloudformation.Tag, len(oldTags))

	// existing tags
	for tagIndex, stackTag := range oldTags {
		if tagValue, exist := (*newTags)[*stackTag.Key]; exist {
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
	for ntKey, ntValue := range *newTags {
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

	return stackTags
}
