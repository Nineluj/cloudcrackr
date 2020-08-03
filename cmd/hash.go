package cmd

import (
	"cloudcrackr/storage"
	"fmt"
	"github.com/spf13/cobra"
)

const HashPrefix = "hash/"

// hashCmd represents the hash command
var hashCmd = &cobra.Command{
	Use:     "hash",
	Aliases: []string{"h"},
	Short:   "Manage the hash files that cloudcrackr has access to",
}

var hashListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Args:    cobra.ExactArgs(0),
	Run:     hashList,
}

var hashAddCmd = &cobra.Command{
	Use:       "add",
	Aliases:   []string{"upload"},
	Short:     "[file] [hash-alias]",
	ValidArgs: nil,
	Args:      cobra.ExactValidArgs(2),
	RunE:      hashAdd,
}

func init() {
	hashCmd.AddCommand(hashAddCmd)
	hashCmd.AddCommand(hashListCmd)
	rootCmd.AddCommand(hashCmd)
}

func hashList(_ *cobra.Command, _ []string) {
	files, err := storage.ListFiles(awsSession, globalCfg.S3BucketName, HashPrefix)

	fmt.Printf("Found a total of [%d] files\n", len(files))
	// Print out the files
	for _, fn := range files {
		fmt.Println(fn)
	}

	_ = err
}

func hashAdd(_ *cobra.Command, args []string) error {
	// Defines the full string that corresponds to the file's key in the S3 bucket
	hashFullKey := HashPrefix + args[1]

	return storage.Upload(awsSession, args[0], globalCfg.S3BucketName, hashFullKey)
}
