package cmd

import (
	"testing"
)

func TestGetRegionsCount(t *testing.T) {
	target := &cfnManager{}

	count := target.getRegionCount()

	if count != 0 {
		t.Error("Failed.")
	}
}

func TestGetRegion(t *testing.T) {
	arn := "arn:partition:service:region:account-id:resourcetype/resource/qualifier"

	region := getRegionFromArn(&arn)

	if region != "region" {
		t.Error("Region incorrect")
	}
}
