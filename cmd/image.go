package cmd

import (
	"cloudcrackr/repository"
	"fmt"

	"github.com/spf13/cobra"
)

// imageCmd represents the image command
var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Handle the images available to cloudcrackr",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("image called")
	},
}

var imagePushCmd = &cobra.Command{
	Use:   "push <imageId> <name>",
	Short: "Push the image ID to the cloudcrackr ECR repository to make it available",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return repository.PushImage(awsSession, args[0], args[1])
	},
}

func init() {
	imageCmd.AddCommand(imagePushCmd)
	rootCmd.AddCommand(imageCmd)
}
