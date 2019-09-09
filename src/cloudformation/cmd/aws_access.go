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
	"github.com/aws/aws-sdk-go/service/ec2"
)

type cfnManagement interface {
	getStack(stackName *string) (*cloudformation.Stack, error)
	getStackTemplate(stackName *string) (*string, error)
	createChangeSet(stackName *string, params []*cloudformation.Parameter, tags []*cloudformation.Tag, templateBody *string, changeSetType string) (*cloudformation.CreateChangeSetOutput, error)
	executeChangeSet(stackname *string, csName *string) error
	getTemplateSummary(templateBody *string) (*cloudformation.GetTemplateSummaryOutput, error)
	getAll(stackChannel chan *cloudformation.Stack, err chan error)
	getRegionCount() int
	delete(stackName *string) error
}

type cfnManager struct {
	cfn             cloudformationiface.CloudFormationAPI
	iamCapabilities []*string
	cfnRegions      map[string]*cloudformationiface.CloudFormationAPI
}

func newCfnClient() *cfnManagement {
	var sess = session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	}))

	ec2client := ec2.New(sess)
	regions, _ := ec2client.DescribeRegions(&ec2.DescribeRegionsInput{AllRegions: aws.Bool(true)})
	cfnPerRegion := make(map[string]*cloudformationiface.CloudFormationAPI)
	for _, region := range regions.Regions {
		if *region.OptInStatus != "opt-in-not-required" {
			continue
		}

		var cfnRegionClient cloudformationiface.CloudFormationAPI = cloudformation.New(session.Must(session.NewSession(&aws.Config{
			Region: region.RegionName,
		})))
		cfnPerRegion[*region.RegionName] = &cfnRegionClient
	}

	var result cfnManagement = &cfnManager{
		cfn: cloudformation.New(sess),
		iamCapabilities: []*string{
			aws.String(cloudformation.CapabilityCapabilityIam),
			aws.String(cloudformation.CapabilityCapabilityNamedIam),
			aws.String(cloudformation.CapabilityCapabilityAutoExpand),
		},
		cfnRegions: cfnPerRegion,
	}
	return &result
}

func (client *cfnManager) delete(stackArn *string) error {
	region := getRegionFromArn(stackArn)
	regionClient := *client.cfnRegions[region]
	dsi := &cloudformation.DeleteStackInput{
		StackName: stackArn,
	}
	_, err := regionClient.DeleteStack(dsi)
	return err
}

func (client *cfnManager) getRegionCount() int {
	return len(client.cfnRegions)
}

func (client *cfnManager) getAll(stackChan chan *cloudformation.Stack, errChan chan error) {
	for _, rc := range client.cfnRegions {
		regionClient := *rc
		go func() {
			rs, rsErr := regionClient.DescribeStacks(&cloudformation.DescribeStacksInput{})
			if rsErr != nil {
				errChan <- rsErr
				return
			}

			for _, stack := range rs.Stacks {
				ss, ssErr := regionClient.DescribeStacks(&cloudformation.DescribeStacksInput{
					StackName: stack.StackName,
				})
				if ssErr != nil {
					errChan <- ssErr
					return
				}
				stackChan <- ss.Stacks[0]
			}

			stackChan <- nil
		}()
	}
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
	return client.cfn.WaitUntilStackUpdateComplete(waitInput)
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

func getRegionFromArn(arn *string) string {
	// arn:partition:service:region:account-id:resource
	segments := strings.Split(*arn, ":")
	return segments[3]
}

func getAccountIdFromArn(arn *string) string {
	// arn:partition:service:region:account-id:resource
	segments := strings.Split(*arn, ":")
	return segments[4]
}
