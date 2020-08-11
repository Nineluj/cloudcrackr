package cmd

import (
	"cloudcrackr/cmd/utility"
	"cloudcrackr/repository"
	"fmt"
	log "github.com/visionmedia/go-cli-log"

	"github.com/spf13/cobra"
)

// imageCmd represents the image command
var imageCmd = &cobra.Command{
	Use:     "image",
	Aliases: []string{"img"},
	Short:   "Handle the images available to cloudcrackr",
}

// Command that creates a new repository if needed (when it doesn't already exist)
// and pushes the image on your local machine
var imagePushCmd = &cobra.Command{
	Use:   "push <imageId> <name>",
	Short: "Push the image ID to the cloudcrackr ECR repository to make it available",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return repository.PushImage(awsSession, args[0], args[1])
	},
}

// The wording on these functions is a bit confusing. Since images are contained in distinct
// repositories (that are in the same registry), imageDelete is deleting the registry and not the
// single image
var imageDeleteCmd = &cobra.Command{
	Use:     "delete <name>",
	Aliases: []string{"rm", "del"},
	Short:   "Deletes the image with the given name",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		err := repository.DeleteImageRepository(awsSession, args[0])
		if err == nil {
			log.Info("Image", "Successfully removed image")
		}
		return err
	},
}

var imageListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List images available for cracking",
	Aliases: []string{"ls"},
	Args:    cobra.ExactArgs(0),
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
	imageCmd.AddCommand(imageDeleteCmd)
	imageCmd.AddCommand(imageListCmd)
	imageCmd.AddCommand(imagePushCmd)
	rootCmd.AddCommand(imageCmd)
}
