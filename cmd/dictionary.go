package cmd

import (
	"cloudcrackr/storage"
	"cloudcrackr/utility"
	"fmt"
	"github.com/spf13/cobra"
	"strings"
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
	Short:   "List dictionaries available for cracking",
	Aliases: []string{"ls"},
	Args:    cobra.ExactArgs(0),
	RunE:    dictionaryList,
}

var dictionaryAddCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"upload"},
	Short:   "[file] [dictionary-alias]",
	Args:    cobra.ExactValidArgs(2),
	RunE:    dictionaryAdd,
}

var dictionaryDeleteCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"rm"},
	Short:   "[dictionary-alias]",
	Args:    cobra.ExactValidArgs(1),
	RunE:    dictionaryDelete,
}

func init() {
	dictionaryCmd.AddCommand(dictionaryAddCmd)
	dictionaryCmd.AddCommand(dictionaryListCmd)
	dictionaryCmd.AddCommand(dictionaryDeleteCmd)
	rootCmd.AddCommand(dictionaryCmd)
}

func dictionaryList(_ *cobra.Command, _ []string) error {
	files, err := storage.ListFiles(awsSession, globalCfg.S3BucketName, DictionaryPrefix)

	if err != nil {
		return err
	}

	fmt.Printf("Found a total of [%d] %v\n", len(files), utility.Pluralize("file", len(files)))
	// Print out the files
	for _, fn := range files {
		fmt.Println("-", strings.TrimLeft(*fn.Key, DictionaryPrefix))
	}

	return nil
}

func dictionaryAdd(_ *cobra.Command, args []string) error {
	// Defines the full string that corresponds to the file's key in the S3 bucket
	dictionaryFullKey := DictionaryPrefix + args[1]

	return storage.Upload(awsSession, args[0], globalCfg.S3BucketName, dictionaryFullKey)
}

func dictionaryDelete(_ *cobra.Command, args []string) error {
	dictionaryFullKey := DictionaryPrefix + args[0]

	return storage.Delete(awsSession, globalCfg.S3BucketName, dictionaryFullKey)
}
