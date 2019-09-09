package cmd

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"testing"
)

func TestDeleteAllCmdPreRunE_ValidatesHasUpdateParams(t *testing.T) {
	var mockCfnManager cfnManagement = &mockCfnManager{}
	ucmd := &deleteAllCmd{
		cm: &CommandManagement{
			cfnManager: &mockCfnManager,
			config:     &config{mode: noninteractive},
		},
	}

	err := ucmd.preRunE(nil, nil)

	if err != nil {
		t.Error("Command delete-all parameters validation failed.")
	}
}

func TestDeleteAllCmdPreRunE_Success(t *testing.T) {
	// arrange
	deletedStacks := make([]*string, 0)
	var mockCfnManager cfnManagement = &mockCfnManager{
		getAllStub: func(stackChan chan *cloudformation.Stack, errChan chan error) {
			stackChan <- &cloudformation.Stack{
				StackName:                   aws.String("a"),
				StackId:                     aws.String("arn:partition:service:region:account-id:resource"),
				StackStatus:                 aws.String(cloudformation.StackStatusCreateComplete),
				EnableTerminationProtection: aws.Bool(false),
			}
			stackChan <- &cloudformation.Stack{
				StackName:                   aws.String("b"),
				EnableTerminationProtection: aws.Bool(true),
			}
			stackChan <- &cloudformation.Stack{
				StackName:                   aws.String("c"),
				StackId:                     aws.String("arn:partition:service:region:account-id:resource"),
				StackStatus:                 aws.String(cloudformation.StackStatusCreateComplete),
				EnableTerminationProtection: aws.Bool(false),
			}
			return
		},
		deleteStub: func(stackArn *string) error {
			deletedStacks = append(deletedStacks, stackArn)
			return nil
		},
		regionCount: 4,
	}
	ucmd := &deleteAllCmd{
		cm: &CommandManagement{
			cfnManager: &mockCfnManager,
			config:     &config{mode: noninteractive},
		},
	}

	err := ucmd.runE(nil, nil)

	if err != nil {
		t.Error("Command delete-all default run .")
	}
	if len(deletedStacks) != 2 {
		fmt.Printf("Stacks:\n%#v\n", deletedStacks)
		t.Error("Incorrect number of stacks")
	}
}
