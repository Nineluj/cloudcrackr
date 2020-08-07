// Handles the deployment of image for cracking passwords
package runner

import (
	"cloudcrackr/compute"
	"cloudcrackr/repository"
	"cloudcrackr/storage"
	"github.com/aws/aws-sdk-go/aws/session"
)

func Crack(sess *session.Session, clusterName, imageName, bucketName, dictionary, hash string, useGpu bool) error {
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
