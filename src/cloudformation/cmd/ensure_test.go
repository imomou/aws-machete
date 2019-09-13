package cmd

import (
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"testing"

	"path"
	"runtime"
	/*"errors"

	  "github.com/aws/aws-sdk-go/aws"
	  "github.com/spf13/viper"*/)

func TestEnsureCmdRunE_CreateSuccess(t *testing.T) {

	// arrange
	_, filename, _, _ := runtime.Caller(0)
	templatePath := path.Join(path.Dir(filename), "..", "main.go")

	mockCfnManager := &mockCfnManager{
		getStackStub: func(stackName *string) (*cloudformation.Stack, error) {
			return nil, nil
		},
	}
	//var cfnManagerInterface cfnManagement = mockCfnManager
	ucmd := &ensureCmd{
		cm: &CommandManagement{
			cfnManager: *mockCfnManager,
			config:     &config{mode: noninteractive},
		},
		templatePath: templatePath,
	}

	// act
	err := ucmd.runE(nil, nil)

	// assert
	if err != nil {
		t.Error("EnsureCmdRun fails to create successfully")
	}
	if len(mockCfnManager.tags) != 0 {
		t.Error("Incorrect number of tags")
	}
	if len(mockCfnManager.params) != 0 {
		t.Error("Incorrect number of parameters")
	}
}

func TestEnsureCmdRunE_UpdateSuccess(t *testing.T) {

	// arrange
	var mockCfnManager cfnManagement = &mockCfnManager{}
	ucmd := &ensureCmd{
		cm: &CommandManagement{
			cfnManager: &mockCfnManager,
			config:     &config{mode: noninteractive},
		},
	}

	// act
	err := ucmd.runE(nil, nil)

	// assert

	if err != nil {
		panic(err)
	}

	if err != nil {
		t.Error("EnsureCmdRun to update successfully")
	}

}

/*
func TestUpdateCmdRunE_SuccessOverridingParam(t *testing.T) {
	// arrange
	var resultParams []*cloudformation.Parameter
	var mockCfnManager cfnManagement = &mockCfnManager{
		getStackStub: func(stackName *string) (*cloudformation.Stack, error) {
			return &cloudformation.Stack{
				Parameters: []*cloudformation.Parameter{
					&cloudformation.Parameter{ // param to modify
						ParameterKey:     aws.String("a"),
						ParameterValue:   aws.String("va"),
						UsePreviousValue: aws.Bool(true), // explicitly set this to test if the value gets correctly set.
					},
					&cloudformation.Parameter{ // param to leave untouched
						ParameterKey:     aws.String("b"),
						ParameterValue:   aws.String("vb"),
						UsePreviousValue: aws.Bool(false), // explicitly set this to test if the value gets correctly set.
					},
					&cloudformation.Parameter{ // param to remove
						ParameterKey:   aws.String("c"),
						ParameterValue: aws.String("vc"),
					},
				},
			}, nil
		},
		createChangeSetStub: func(stackName *string, params []*cloudformation.Parameter, tags []*cloudformation.Tag, templateBody *string, changeSetType string) (*cloudformation.CreateChangeSetOutput, error) {
			resultParams = params
			return &cloudformation.CreateChangeSetOutput{}, nil
		},
		getTemplateSummaryStub: func(templateBody *string) (*cloudformation.GetTemplateSummaryOutput, error) {
			return &cloudformation.GetTemplateSummaryOutput{
				Parameters: []*cloudformation.ParameterDeclaration{
					&cloudformation.ParameterDeclaration{
						ParameterKey: aws.String("a"),
					},
					&cloudformation.ParameterDeclaration{
						ParameterKey: aws.String("b"),
					},
					&cloudformation.ParameterDeclaration{ // param to add
						ParameterKey: aws.String("d"),
					},
				},
			}, nil
		},
	}
	ucmd := &updateCmd{
		cm: &CommandManagement{
			cfnManager: &mockCfnManager,
			config:     &config{mode: noninteractive},
		},
		params: map[string]string{
			"a": "vaa",
			"d": "vdd",
		},
	}

	// act
	err := ucmd.runE(nil, nil)

	fmt.Printf("%#v", resultParams)

	// assert
	if err != nil {
		t.Error("UpdateCmdRun should not fail.")
	}
	if len(resultParams) != 3 {
		t.Error("Incorrect number of tags")
	}
	for _, param := range resultParams {
		switch *param.ParameterKey {
		case "a":
			if *param.ParameterValue != "vaa" {
				t.Error("Tag overriding failed")
			}
			if *param.UsePreviousValue {
				t.Error("Tag a should not be using previous value.")
			}
		case "b":
			if !*param.UsePreviousValue {
				t.Error("Tag b should be using previous value.")
			}
		case "d":
			if *param.ParameterValue != "vdd" {
				t.Error("Added tag value wrong.")
			}
			if *param.UsePreviousValue {
				t.Error("Tag d should not be using previous value.")
			}
		default:
			t.Errorf("This parameter shouldn't exist.\n%#v", param)
		}
	}
}

func TestUpdateCmdRunE_SuccessOverridingTag(t *testing.T) {
	// arrange
	var resultTags []*cloudformation.Tag
	var mockCfnManager cfnManagement = &mockCfnManager{
		getStackStub: func(stackName *string) (*cloudformation.Stack, error) {
			return &cloudformation.Stack{
				Tags: []*cloudformation.Tag{
					&cloudformation.Tag{Key: aws.String("a"), Value: aws.String("va")},
					&cloudformation.Tag{Key: aws.String("b"), Value: aws.String("vb")},
				},
			}, nil
		},
		createChangeSetStub: func(stackName *string, params []*cloudformation.Parameter, tags []*cloudformation.Tag, templateBody *string, changeSetType string) (*cloudformation.CreateChangeSetOutput, error) {
			resultTags = tags
			return &cloudformation.CreateChangeSetOutput{}, nil
		},
	}
	ucmd := &updateCmd{
		cm: &CommandManagement{
			cfnManager: &mockCfnManager,
			config:     &config{mode: noninteractive},
		},
		tags: map[string]string{
			"a": "vaa",
			"c": "vcc",
		},
	}

	// act
	err := ucmd.runE(nil, nil)

	// assert
	if err != nil {
		t.Error("UpdateCmdRun should not fail.")
	}
	if len(resultTags) != 3 {
		t.Error("Incorrect number of tags")
	}
	for _, tag := range resultTags {
		switch *tag.Key {
		case "a":
			if *tag.Value != "vaa" {
				t.Error("Tag overriding failed")
			}
		case "b":
			if *tag.Value != "vb" {
				t.Error("Tag changed without being specified")
			}
		case "c":
			if *tag.Value != "vcc" {
				t.Error("Added tag value wrong.")
			}
		default:
			t.Errorf("This tag shouldn't exist.\n%#v", tag)
		}
	}
}

func TestUpdateCmdRunE_StackNotExist(t *testing.T) {
	// Setup
	var mockCfnManager cfnManagement = &mockCfnManager{
		getStackStub: func(stackName *string) (*cloudformation.Stack, error) {
			return nil, errors.New("stack does not exist")
		},
	}
	ucmd := &updateCmd{
		cm: &CommandManagement{
			cfnManager: &mockCfnManager,
		},
	}

	err := ucmd.runE(nil, nil)

	if err == nil {
		t.Error("Should fail when error returned from getStack")
	}
}
*/
