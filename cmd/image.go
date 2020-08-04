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
	Use:  "push",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return repository.PushImage(awsSession, args[0])
	},
}

func init() {
	imageCmd.AddCommand(imagePushCmd)
	rootCmd.AddCommand(imageCmd)
}
