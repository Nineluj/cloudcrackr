package cmd

import (
	"cloudcrackr/utility"
	"github.com/spf13/cobra"
)

// teardownCmd represents the teardown command
var teardownCmd = &cobra.Command{
	Use:   "teardown",
	Short: "Tear down the existing infrastructure",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: tearDown,
}

var force bool

func tearDown(cmd *cobra.Command, args []string) {
	if !force {
		accept := utility.GetBoolean("This will remove the existing infrastructure. " +
			"This operation cannot be reversed. Proceed?")

		if !accept {
			return
		}
	}

	// TODO: remove S3 bucket, ECS images & containers, ...
}

func init() {
	rootCmd.AddCommand(teardownCmd)
	rootCmd.Flags().BoolVar(&force, "force", false, "used to force teardown, avoid prompt")
}
