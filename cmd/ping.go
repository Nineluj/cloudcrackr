package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"strconv"
)

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Useless command, to be deleted!",
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

	return nil
}

func init() {
	rootCmd.AddCommand(pingCmd)
}
