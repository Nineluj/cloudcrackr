// This project is not likely to use dynamodb anymore, keeping this here for now anyway
package dynamodb

//
//import (
//	"fmt"
//	"github.com/aws/aws-sdk-go/aws"
//	"github.com/aws/aws-sdk-go/aws/session"
//	"github.com/aws/aws-sdk-go/service/dynamodb"
//	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
//	"github.com/visionmedia/go-cli-log"
//)
//
//
//func ListEntries(tableName string, awsSession *session.Session) error {
//	var passwordFileBuckets []S3Entry
//
//	ddbClient := dynamodb.New(awsSession)
//
//	result, err := ddbClient.Scan(&dynamodb.ScanInput{
//		TableName: aws.String(tableName),
//	})
//
//	if err != nil {
//		return err
//	}
//
//	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &passwordFileBuckets)
//
//	if err != nil {
//		return err
//	}
//
//	log.Info("list", "Items in %s", tableName)
//
//	for pos, item := range passwordFileBuckets {
//		fmt.Printf(
//			"[%d/%d] Alias={%s} Bucket={%s} Key={%s}\n",
//			pos+1,
//			len(passwordFileBuckets),
//			item.Alias,
//			item.BucketName,
//			item.KeyName,
//		)
//	}
//
//	return nil
//}

//func AddEntry(tableName string, args cli.Args, awsSession *session.Session) error {
//	if len(args) != 3 {
//		return cli.NewExitError("Invalid usage, need <alias> <bucketName> <bucketKey>", 1)
//	}
//
//	newEntry := models.S3Entry{
//		Alias:      args.Get(0),
//		BucketName: args.Get(1),
//		KeyName:    args.Get(2),
//	}
//
//	av, err := dynamodbattribute.MarshalMap(newEntry)
//
//	if err != nil {
//		return err
//	}
//
//	ddbClient := dynamodb.New(awsSession)
//
//	_, err = ddbClient.PutItem(&dynamodb.PutItemInput{
//		TableName: aws.String(tableName),
//		Item:      av,
//	})
//
//	if err != nil {
//		return err
//	}
//
//	ccrlog.Info(fmt.Sprintf("Successfully added new entry in %s", tableName))
//
//	return nil
//}
//
//func RemoveEntry(tableName string, args cli.Args, awsSession *session.Session) error {
//	if len(args) != 1 {
//		return cli.NewExitError("Invalid usage, need <alias>", 1)
//	}
//
//	ddbClient := dynamodb.New(awsSession)
//
//	_, err := ddbClient.DeleteItem(&dynamodb.DeleteItemInput{
//		Key: map[string]*dynamodb.AttributeValue{
//			"Alias": {
//				N: aws.String(args.Get(0)),
//			},
//		},
//		TableName: aws.String(tableName),
//	})
//
//	if err != nil {
//		return err
//	}
//
//	ccrlog.Info("Successfully removed password option. Did not change the associated S3 bucket.")
//
//	return nil
//}
