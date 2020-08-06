// Provides functions for interacting with storage on Cloud
package storage

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	log "github.com/visionmedia/go-cli-log"
	"os"
)

// Initiates the storage on AWS that cloudcrackr can use
func CreateBucket(sess *session.Session, bucketName string) error {
	client := s3.New(sess)

	// If the bucket exist no error is thrown
	_, err := client.CreateBucket(
		&s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
		},
	)

	if err != nil {
		return err
	}

	return nil
}

// Lists the files available with the given prefix.
// Useful to retrieve only the password or hash files
func ListFiles(sess *session.Session, bucketName, prefix string) ([]*s3.Object, error) {
	client := s3.New(sess)

	result, err := client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(prefix),
	})

	if err != nil {
		return nil, err
	}

	return result.Contents, nil
}

// Upload a file to the remote storage from the local storage.
// Key represents the location of the uploaded file in the remote storage.
func Upload(sess *session.Session, filePath, bucketName, key string) error {
	uploader := s3manager.NewUploader(sess)

	f, err := os.Open(filePath)

	if err != nil {
		return err
	}

	// Upload the file to S3.
	// could use custom reader to add progress info:
	// https://github.com/aws/aws-sdk-go/blob/master/example/service/s3/putObjectWithProcess/putObjWithProcess.go
	log.Info("Upload", "uploading file")
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   f,
	})

	if err != nil {
		return err
	}

	log.Info("Upload", "File successfully uploaded")

	return nil
}

// Delete a file from the remote storage
func Delete(sess *session.Session, bucketName, key string) error {
	client := s3.New(sess)

	_, err := client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})

	return err
}

// Stat checks if the file name is present
func stat(client *s3.S3, bucketName, key string, errChan chan<- error) {
	// Getting ACL is faster than getting the whole object
	_, err := client.GetObjectAcl(&s3.GetObjectAclInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})

	errChan <- err
}

func Stat(sess *session.Session, bucketName, key string) error {
	client := s3.New(sess)
	errChan := make(chan error)

	go stat(client, bucketName, key, errChan)
	return <-errChan
}

func StatMultiple(sess *session.Session, bucketName string, keyList ...string) error {
	keysLen := len(keyList)
	client := s3.New(sess)
	errChan := make(chan error)

	for _, key := range keyList {
		go stat(client, bucketName, key, errChan)
	}

	for i := 0; i < keysLen; i++ {
		res := <-errChan
		if res != nil {
			return res
		}
	}

	return nil
}
