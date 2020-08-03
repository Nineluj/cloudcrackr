package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Useless command, to be deleted!",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Pong!")
	},
}

func init() {
	rootCmd.AddCommand(pingCmd)
}
