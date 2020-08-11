package cmd

import (
	"cloudcrackr/auth"
	"cloudcrackr/compute"
	"cloudcrackr/storage"
	"github.com/spf13/cobra"
	log "github.com/visionmedia/go-cli-log"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use: "init",
	Short: "Create the infrastructure necessary for running cloudcrackr. The command is idempotent," +
		" ie can be run multiple times without any changes or errors",
	RunE: initInfra,
}

func initInfra(_ *cobra.Command, _ []string) error {
	err := storage.CreateBucket(awsSession, globalCfg.S3BucketName)
	if err != nil {
		return err
	}

	// This is not needed: repositories will be created when a new image is added
	//err = repository.CreateRepository(awsSession)
	//if err != nil {
	//	return err
	//}
	//

	err = compute.CreateCluster(awsSession, globalCfg.ClusterName)
	if err != nil {
		return err
	}

	err = auth.SetupIAM(awsSession, globalCfg.IAMRoleNamePath)
	if err != nil {
		return err
	}

	log.Info("Initialization", "Complete")

	return nil
}

func init() {
	rootCmd.AddCommand(initCmd)
}
