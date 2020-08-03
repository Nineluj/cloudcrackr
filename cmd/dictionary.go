package cmd

import (
	"cloudcrackr/storage"
	"fmt"
	"github.com/spf13/cobra"
)

const DictionaryPrefix = "dictionary/"

// dictionaryCmd represents the dictionary command
var dictionaryCmd = &cobra.Command{
	Use:     "dictionary",
	Aliases: []string{"dict", "d", "pass"},
	Short:   "Manage the dictionaries that cloudcrackr has access to",
}

var dictionaryListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Args:    cobra.ExactArgs(0),
	Run:     dictionaryList,
}

var dictionaryAddCmd = &cobra.Command{
	Use:       "add",
	Aliases:   []string{"upload"},
	Short:     "[file] [dictionary-alias]",
	ValidArgs: nil,
	Args:      cobra.ExactValidArgs(2),
	RunE:      dictionaryAdd,
}

func init() {
	dictionaryCmd.AddCommand(dictionaryAddCmd)
	dictionaryCmd.AddCommand(dictionaryListCmd)
	rootCmd.AddCommand(dictionaryCmd)
}

func dictionaryList(_ *cobra.Command, _ []string) {
	files, err := storage.ListFiles(awsSession, globalCfg.S3BucketName, DictionaryPrefix)

	fmt.Printf("Found a total of [%d] files\n", len(files))
	// Print out the files
	for _, fn := range files {
		fmt.Println(fn)
	}

	_ = err
}

func dictionaryAdd(_ *cobra.Command, args []string) error {
	// Defines the full string that corresponds to the file's key in the S3 bucket
	dictionaryFullKey := DictionaryPrefix + args[1]

	return storage.Upload(awsSession, args[0], globalCfg.S3BucketName, dictionaryFullKey)
}
