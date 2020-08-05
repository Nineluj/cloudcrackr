package cmd

import (
	"cloudcrackr/repository"
	"cloudcrackr/utility"
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

var imageListCmd = &cobra.Command{
	Use:   "list",
	Short: "List images available for cracking",
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		images, err := repository.ListImages(awsSession)

		imagesLen := len(images)

		fmt.Printf("Found [%d] %v\n", imagesLen, utility.Pluralize("image", imagesLen))
		for _, img := range images {
			fmt.Printf("- %v\n", img)
		}

		return err
	},
}

func init() {
	imageCmd.AddCommand(imageListCmd)
	imageCmd.AddCommand(imagePushCmd)
	rootCmd.AddCommand(imageCmd)
}
