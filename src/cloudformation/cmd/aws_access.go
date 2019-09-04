package cmd

import (
	"errors"
	"fmt"
	"github.com/nu7hatch/gouuid"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

type cfnManagement interface {
	getStack(stackName *string) (*cloudformation.Stack, error)
	getStackTemplate(stackName *string) (*string, error)
	createChangeSet(stackName *string, params []*cloudformation.Parameter, tags []*cloudformation.Tag, templateBody *string, changeSetType string) (*cloudformation.CreateChangeSetOutput, error)
	executeChangeSet(stackname *string, csName *string) error
	getTemplateSummary(templateBody *string) (*cloudformation.GetTemplateSummaryOutput, error)
}

type cfnManager struct {
	cfn             cloudformationiface.CloudFormationAPI
	iamCapabilities []*string
}

func newCfnClient() *cfnManagement {
	var sess = session.Must(session.NewSession())

	var result cfnManagement = &cfnManager{
		cfn: cloudformation.New(sess),
		iamCapabilities: []*string{
			aws.String(cloudformation.CapabilityCapabilityIam),
			aws.String(cloudformation.CapabilityCapabilityNamedIam),
			aws.String(cloudformation.CapabilityCapabilityAutoExpand),
		},
	}
	return &result
}

func (client *cfnManager) createChangeSet(
	stackName *string,
	params []*cloudformation.Parameter,
	tags []*cloudformation.Tag,
	templateBody *string,
	changeSetType string) (*cloudformation.CreateChangeSetOutput, error) {

	guid, guidErr := uuid.NewV4()
	if guidErr != nil {
		return nil, guidErr
	}
	guidString := "ChangeSet-" + strings.Split(guid.String(), "-")[4]

	usePreviousTemplate := templateBody == nil

	csInput := &cloudformation.CreateChangeSetInput{
		StackName:           stackName,
		Capabilities:        client.iamCapabilities,
		Parameters:          params,
		Tags:                tags,
		TemplateBody:        templateBody,
		ChangeSetName:       &guidString,
		ChangeSetType:       &changeSetType,
		UsePreviousTemplate: &usePreviousTemplate,
	}

	// Create change set.
	result, changeSetErr := client.cfn.CreateChangeSet(csInput)
	if changeSetErr != nil {
		return nil, changeSetErr
	}

	// Wait for change set to finish creating.
	waitInput := &cloudformation.DescribeChangeSetInput{
		ChangeSetName: result.Id,
		StackName:     result.StackId,
	}
	waitErr := client.cfn.WaitUntilChangeSetCreateComplete(waitInput)
	if changeSetErr != nil {
		return nil, waitErr
	}

	return result, nil
}

func (client *cfnManager) executeChangeSet(stackname *string, csName *string) error {

	// Execute changeset.
	ecsInput := &cloudformation.ExecuteChangeSetInput{
		ChangeSetName: csName,
		StackName:     stackname,
	}
	_, ecsErr := client.cfn.ExecuteChangeSet(ecsInput)
	if ecsErr != nil {
		return ecsErr
	}

	// Wait changeset to finishe executing.
	waitInput := &cloudformation.DescribeStacksInput{
		StackName: stackname,
	}
	waitErr := client.cfn.WaitUntilStackUpdateComplete(waitInput)
	if waitErr != nil {
		return waitErr
	}

	return nil
}

func (client *cfnManager) getTemplateSummary(templateBody *string) (*cloudformation.GetTemplateSummaryOutput, error) {
	gtsInput := &cloudformation.GetTemplateSummaryInput{
		TemplateBody: templateBody,
	}
	return client.cfn.GetTemplateSummary(gtsInput)
}

func (client *cfnManager) getStack(stackName *string) (*cloudformation.Stack, error) {
	result, err := client.cfn.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: stackName,
	})

	if err != nil {
		return nil, err
	}

	if len(result.Stacks) != 1 {
		return nil, errors.New(fmt.Sprintf("Multiple stacks found.\n\n%#v", result.Stacks))
	}

	return result.Stacks[0], nil
}

func (client *cfnManager) getStackTemplate(stackName *string) (*string, error) {
	result, err := client.cfn.GetTemplate(&cloudformation.GetTemplateInput{
		StackName: stackName,
	})

	if err != nil {
		return nil, err
	}

	return result.TemplateBody, nil
}
