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
	result, err := client.CreateBucket(
		&s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
		},
	)

	if err != nil {
		return err
	}

	_ = result

	return nil
}

func ListFiles(sess *session.Session, bucketName, prefix string) ([]s3.Object, error) {
	client := s3.New(sess)

	result, err := client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(prefix),
	})

	if err != nil {
		return nil, err
	}

	fmt.Println(result)

	return nil, nil
}

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

func Delete(sess *session.Session, bucketName, key string) error {
	client := s3.New(sess)

	_, err := client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})

	return err
}
