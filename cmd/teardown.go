package cmd

import (
	"cloudcrackr/auth"
	"cloudcrackr/cmd/utility"
	"cloudcrackr/compute"
	"cloudcrackr/storage"
	"github.com/spf13/cobra"
)

// teardownCmd represents the teardown command
var teardownCmd = &cobra.Command{
	Use:   "teardown",
	Short: "Tear down the existing infrastructure. Use conf clean to remove the configuration file.",
	RunE:  tearDown,
}

var force bool

func tearDown(_ *cobra.Command, _ []string) error {
	if !force {
		accept := utility.GetBoolean("This will remove the existing infrastructure. " +
			"This operation cannot be reversed. Proceed?")
		if !accept {
			return nil
		}
	}

	// Remove the S3 bucket
	err := storage.DeleteBucket(awsSession, globalCfg.S3BucketName)
	if err != nil {
		return err
	}

	// remove ECS cluster
	err = compute.DeleteCluster(awsSession, globalCfg.ClusterName)
	if err != nil {
		return err
	}

	// Remove IAM image
	return auth.DeleteIAMRoles(awsSession)
}

func init() {
	rootCmd.AddCommand(teardownCmd)
	rootCmd.Flags().BoolVar(&force, "force", false, "used to force teardown, avoid prompt")
}
