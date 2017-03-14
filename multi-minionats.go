package main


import (
	// "io"
	// "log"
	"os"
	"fmt"
	"runtime"
	"encoding/json"

	"github.com/minio/minio-go"
	nats "github.com/nats-io/go-nats"
)

func printBuckets(s3client  minio.Client) {
	buckets, err := s3client.ListBuckets()
	if err != nil {
		fmt.Printf("error listing buckets: %v\n", err)
	}
	for _, bucket := range buckets {
		fmt.Printf("found bucket: %v\n", bucket.Name)
	}
}

func main() {

	// assumes:
	// 	1. Running minio server.
	//	2. Running natsd server.
	// open connection to remote s3 bucket.
	bucket := "minio-nats-example"
	location := "us-east-1"

	remoteS3AccessKeyId := os.Getenv("MINIO_AWS_ACCESS_KEY_ID")
	remoteS3SecretKey := os.Getenv("MINIO_AWS_SECRET_ACCESS_KEY")

	s3RemoteClient, err := minio.New("s3.amazonaws.com", remoteS3AccessKeyId,
		remoteS3SecretKey, true)

	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	// open connection to local s3 bucket.
	printBuckets(*s3RemoteClient)

	s3LocalClient, err := minio.New("10.0.1.17:9000", "9W2392IOEBUAZH6PLCHA",
		"g7XPw7gWdVoazRLoVfg1ZcV3SrcUSz3gwhUonTVC", false)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	// add notifications to local s3 bucket.
	printBuckets(*s3LocalClient)

	// create the local bucket if it doesn't exist
	exists, err := s3LocalClient.BucketExists(bucket)
	if err != nil {
		fmt.Printf("error checking bucket exists: %v", err)
	}

	if !exists {
		s3LocalClient.MakeBucket(bucket, location)
	}

	// check for matching remote bucket
	exists, err = s3RemoteClient.BucketExists(bucket)
	if err != nil {
		fmt.Printf("error checking bucket exsists: %v", err)
	}

	if !exists {
		s3RemoteClient.MakeBucket(bucket, location)
	}

	//// on bucket notification
	//topicArn := minio.NewArn("minio", "sqs", location, "1", "nats")
	//topicConfig := minio.NewNotificationConfig(topicArn)
	//topicConfig.AddEvents(minio.ObjectCreatedAll, minio.ObjectRemovedAll)
	//topicConfig.AddFilterSuffix(".jpg")
	//
	//bucketNotification := minio.BucketNotification{}
	//bucketNotification.AddTopic(topicConfig)
	//err = s3LocalClient.SetBucketNotification(bucket, bucketNotification)
	//if err != nil {
	//	fmt.Printf("Unable to set bucket notification %v\n", err)
	//}
	//
	//// get object from local s3.
	//bucketNotification2, err := s3LocalClient.GetBucketNotification(bucket)
	//fmt.Printf("bucketNotification2: %v", bucketNotification2)
	//
	//if err != nil {
	//	fmt.Printf("failed to get bucket notifications for %s, %v\n", bucket, err)
	//}
	//
	//for _, topicConfig := range bucketNotification2.TopicConfigs {
	//	for _, e := range topicConfig.Events {
	//		fmt.Println(e + " event is enabled")
	//	}
	//}

	natsConnection, _ := nats.Connect("nats://localhost:4222")
	fmt.Println("Connected")

	// Subscribe to subject
	fmt.Printf("Subscribing to subject 'bucketevents'\n")
	natsConnection.Subscribe("bucketevents", func(msg *nats.Msg) {

		// Handle the message
		fmt.Printf("Received message '%s\n", string(msg.Data) + "'")

		notification := minio.NotificationInfo{}

		err = json.Unmarshal(msg.Data, &notification)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		}

		for _, record := range notification.Records {
			fmt.Printf("event %v\n", record.S3.Object.Key)
			if record.EventName == "s3:ObjectCreated:Put" {
				fmt.Printf("we need to download, then copy this object to the other bucket.\n")
				err = s3LocalClient.FGetObject(bucket, record.S3.Object.Key, fmt.Sprintf("/tmp/%s/%s", bucket, record.S3.Object.Key))
				if err != nil {
					fmt.Printf("error: %v\n", err)
				}
				_, err := s3RemoteClient.FPutObject(bucket, record.S3.Object.Key, fmt.Sprintf("/tmp/%s/%s", bucket, record.S3.Object.Key), "application/octet-stream")
				if err != nil {
					fmt.Printf("error: %v\n", err)
				}
			}
		}

		// get the object from the local client

	})

	// Keep the connection alive
	runtime.Goexit()

	// transmit object to remote s3.

}