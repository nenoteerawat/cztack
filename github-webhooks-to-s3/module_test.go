package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
)

func TestGitHubWebhooks(t *testing.T) {
	options := &terraform.Options{
		TerraformDir: ".",
	}
	terraform.InitAndPlan(t, options)
}
