// Handles the deployment of image for cracking passwords
package runner

import (
	"cloudcrackr/compute"
	"cloudcrackr/repository"
	"errors"
	"github.com/aws/aws-sdk-go/aws/session"
)

func crack(sess *session.Session, imageName string) error {
	// Retrieve info about image
	imageURI, err := repository.GetImageURI(sess, imageName)
	if err != nil {
		return err
	}
	// check image validity

	// Set up Storage Gateway for session

	// Deploy image
	// TODO: complete this
	err = compute.DeployContainer(sess, imageURI, false)
	if err != nil {
		return err
	}

	return errors.New("not implemented")
}
