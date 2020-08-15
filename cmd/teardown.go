package cmd

import (
	"cloudcrackr/cmd/utility"
	"errors"
	"github.com/spf13/cobra"
)

// teardownCmd represents the teardown command
var teardownCmd = &cobra.Command{
	Use:   "teardown",
	Short: "Tear down the existing infrastructure. Use conf clean to remove the configuration file.",
	RunE:  tearDown,
}

var force bool

func tearDown(cmd *cobra.Command, args []string) error {
	if !force {
		accept := utility.GetBoolean("This will remove the existing infrastructure. " +
			"This operation cannot be reversed. Proceed?")
		if !accept {
			return nil
		}
	}

	// TODO: remove S3 bucket, ECS images & containers, IAM role last...
	return errors.New("not implemented")
}

func init() {
	rootCmd.AddCommand(teardownCmd)
	rootCmd.Flags().BoolVar(&force, "force", false, "used to force teardown, avoid prompt")
}
