package storage

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	log "github.com/visionmedia/go-cli-log"
	"os"
)

func New(sess *session.Session, bucketName string) error {
	client := s3.New(sess)

	// TODO: handle duplicate name error more gracefully
	output, err := client.CreateBucket(
		&s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
		},
	)

	if err != nil {
		return err
	}

	_ = output

	return nil
}

func upload(sess *session.Session, key, filePath, bucketName string) error {
	uploader := s3manager.NewUploader(sess)

	f, err := os.Open(filePath)

	if err != nil {
		return err
	}

	// Upload the file to S3.
	log.Info("Upload", "uploading file")
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   f,
	})

	if err != nil {
		return err
	}

	fmt.Printf("file uploaded to, %s\n", result.Location)

	return nil
}
