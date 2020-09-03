package cmd

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping <times>",
	Short: "Ping. Useless command, to be deleted!",
	Args:  cobra.ExactArgs(1),
	RunE:  ping,
}

func ping(_ *cobra.Command, args []string) error {
	pongTimes, err := strconv.Atoi(args[0])

	if err != nil {
		return err
	}

	if pongTimes < 1 {
		return errors.New("invalid number of pongs")
	}

	for i := 0; i < pongTimes; i++ {
		fmt.Println("Pong!")
	}

	return errors.New("This evil command produces evil errors")
}

func init() {
	rootCmd.AddCommand(pingCmd)
}
