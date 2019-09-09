package cmd

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

type mockCfnManager struct {
	params                 []*cloudformation.Parameter
	tags                   []*cloudformation.Tag
	getStackStub           func(stackName *string) (*cloudformation.Stack, error)
	getStackTemplateStub   func(stackName *string) (*string, error)
	createChangeSetStub    func(stackName *string, params []*cloudformation.Parameter, tags []*cloudformation.Tag, templateBody *string, changeSetType string) (*cloudformation.CreateChangeSetOutput, error)
	executeChangeSetStub   func(stackname *string, csName *string) error
	getTemplateSummaryStub func(templateBody *string) (*cloudformation.GetTemplateSummaryOutput, error)
	getAllStub             func(stackChannel chan *cloudformation.Stack, errChannel chan error)
	deleteStub             func(stackName *string) error
	regionCount            int
}

func (mcm *mockCfnManager) getStack(stackName *string) (*cloudformation.Stack, error) {
	if mcm.getStackStub == nil {
		return &cloudformation.Stack{}, nil
	}
	return mcm.getStackStub(stackName)
}

func (mcm *mockCfnManager) getStackTemplate(stackName *string) (*string, error) {
	if mcm.getStackTemplateStub == nil {
		return aws.String(""), nil
	}
	return mcm.getStackTemplateStub(stackName)
}

func (mcm *mockCfnManager) createChangeSet(stackName *string, params []*cloudformation.Parameter, tags []*cloudformation.Tag, templateBody *string, changeSetType string) (*cloudformation.CreateChangeSetOutput, error) {
	mcm.tags = tags
	mcm.params = params
	if mcm.createChangeSetStub == nil {
		return &cloudformation.CreateChangeSetOutput{}, nil
	}
	return mcm.createChangeSetStub(stackName, params, tags, templateBody, changeSetType)
}

func (mcm *mockCfnManager) executeChangeSet(stackname *string, csName *string) error {
	if mcm.executeChangeSetStub == nil {
		return nil
	}
	return mcm.executeChangeSetStub(stackname, csName)
}

func (mcm *mockCfnManager) getTemplateSummary(templateBody *string) (*cloudformation.GetTemplateSummaryOutput, error) {
	if mcm.getTemplateSummaryStub == nil {
		return &cloudformation.GetTemplateSummaryOutput{}, nil
	}
	return mcm.getTemplateSummaryStub(templateBody)
}

func (mcm *mockCfnManager) getAll(stackChannel chan *cloudformation.Stack, errChannel chan error) {
	if mcm.getAllStub != nil {
		mcm.getAllStub(stackChannel, errChannel)
	}

	for i := 0; i < mcm.regionCount; i++ {
		stackChannel <- nil
	}
}

func (mcm *mockCfnManager) getRegionCount() int {
	return mcm.regionCount
}

func (mcm *mockCfnManager) delete(stackArn *string) error {
	if mcm.deleteStub != nil {
		return mcm.deleteStub(stackArn)
	}

	return nil
}
