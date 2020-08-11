package cmd

import (
	"cloudcrackr/compute"
	"cloudcrackr/repository"
	"cloudcrackr/storage"
	"github.com/aws/aws-sdk-go/aws/session"
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

		return crack(
			awsSession,
			globalCfg.ClusterName,
			imageName,
			globalCfg.S3BucketName,
			DictionaryPrefix+dictionary,
			HashPrefix+hash,
			useGpu,
		)
	},
}

func crack(sess *session.Session, clusterName, imageName, bucketName, dictionary, hash string, useGpu bool) error {
	// Retrieve info about image
	imageURI, err := repository.GetImageURI(sess, imageName)
	if err != nil {
		return err
	}

	// Check for the presence of the dictionary and hash file
	err = storage.StatMultiple(sess, bucketName, dictionary, hash)
	if err != nil {
		return err
	}

	// Deploy image
	err = compute.DeployContainer(sess, clusterName, imageURI, bucketName, dictionary, hash, useGpu)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(crackCmd)

	crackCmd.Flags().BoolVar(&useGpu, "use-gpu",
		false, "Use a GPU for the cracking")

	// Haven't decided on behavior: should hash file be deleted after?
	crackCmd.Flags().BoolVar(&uploadHash, "local-hash",
		false, "upload the hash file specified")
}
