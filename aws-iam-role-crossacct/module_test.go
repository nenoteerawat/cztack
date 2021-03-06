package test

import (
	"testing"

	"github.com/chanzuckerberg/cztack/testutil"
	"github.com/gruntwork-io/terratest/modules/random"
)

func TestAWSIAMRoleCrossAcct(t *testing.T) {

	curAcct := testutil.AWSCurrentAccountId(t)

	terraformOptions := testutil.Options(
		testutil.IAMRegion,

		map[string]interface{}{
			"role_name":         random.UniqueId(),
			"source_account_id": curAcct,
			"role_tags": map[string]string{
				"test": random.UniqueId(),
			},
		},
	)

	defer testutil.Cleanup(t, terraformOptions)

	testutil.Run(t, terraformOptions)
}
