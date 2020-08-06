// Handles the deployment of image for cracking passwords
package runner

import (
	"cloudcrackr/compute"
	"errors"
	"github.com/aws/aws-sdk-go/aws/session"
)

func crack(sess *session.Session, imageName string) error {
	// Retrieve info about image

	// check image validity

	// Set up Storage Gateway for session

	// Deploy image
	// TODO: complete this
	compute.DeployContainer(sess)

	return errors.New("not implemented")
}
