package cmd

import (
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"testing"

	"path"
	"runtime"
)

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
			cfnManager: mockCfnManager,
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
	ucmd := &ensureCmd{
		cm: &CommandManagement{
			cfnManager: &mockCfnManager{},
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
