package dynamodb

type S3Entry struct {
	Alias      string `json:"alias"`
	BucketName string `json:"bucketName"`
	KeyName    string `json:"keyName"`
}
