package cmd

import (
	"cloudcrackr/runner"
	"github.com/spf13/cobra"
	"path/filepath"
)

var (
	useGpu, uploadHash bool
)

// crackCmd represents the crack command
var crackCmd = &cobra.Command{
	Use:   "crack <image> <password> <file>",
	Short: "Launches the cracking instance",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		imageName := args[0]
		dictionary := args[1]
		hash := args[2]

		if uploadHash {
			hashFileBase := filepath.Base(hash)
			err := hashAdd(&cobra.Command{}, []string{hash, hashFileBase})
			if err != nil {
				return err
			}

			hash = hashFileBase
		}

		return runner.Crack(
			awsSession,
			imageName,
			globalCfg.S3BucketName,
			DictionaryPrefix+dictionary,
			HashPrefix+hash,
			useGpu,
		)
	},
}

func init() {
	rootCmd.AddCommand(crackCmd)

	crackCmd.Flags().BoolVar(&useGpu, "use-gpu",
		false, "Use a GPU for the cracking")

	// Haven't decided on behavior: should hash file be deleted after?
	crackCmd.Flags().BoolVar(&uploadHash, "local-hash",
		false, "upload the hash file specified")
}
