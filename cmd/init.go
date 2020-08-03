package cmd

import (
	"cloudcrackr/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create the infrastructure necessary for running ",
	RunE:  initInfra,
}

func initInfra(cmd *cobra.Command, args []string) error {
	err := storage.New(awsSession, viper.GetString("S3BucketName"))
	if err != nil {
		return err
	}

	//

	return nil
}

func init() {
	rootCmd.AddCommand(initCmd)
}
