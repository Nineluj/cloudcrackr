package cmd

import (
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

func init() {
	rootCmd.AddCommand(imageCmd)
}
