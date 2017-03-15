package main


import (
	// "io"
	"log"
	"os"
	"fmt"
	"runtime"
	"encoding/json"
	"net/url"

	"github.com/minio/minio-go"
	nats "github.com/nats-io/go-nats"
)

func printMinion() {
	log.Print(`Starting MinioNATS
	         ,_---~~~~~----._
	  _,,_,*^____      _____'''*g*\"*,                      MinioNATS Demo
	 / __/ /'     ^.  /      \ ^@q   f                      Peter Miron
	[  @f | @))    |  | @))   l  0 _/                       @petemiron
	 \ /   \~____ / __ \_____/    \
	 |           _l__l_           I                         @nats_io
	 }          [______]           I                        nats.io
	 ]            | | |            |
	 ]             ~ ~             |
	  |                            |
	  |                           |
	───▐▓▓▌═════════════════════▐▓▓▌
	───▐▐▓▓▌▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▐▓▓▌▌
	───█══▐▓▄▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▄▓▌══█
	──█══▌═▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▌═▐══█
	──█══█═▐▓▓▓▓▓▓▄▄▄▄▄▄▄▓▓▓▓▓▓▌═█══█
	──█══█═▐▓▓▓▓▓▓▐██▀██▌▓▓▓▓▓▓▌═█══█
	──█══█═▐▓▓▓▓▓▓▓▀▀▀▀▀▓▓▓▓▓▓▓▌═█══█
	──█══█▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓█══█
	─▄█══█▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▌█══█▄
	─█████▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▌─█████
	─██████▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▌─██████
	──▀█▀█──▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▌───█▀█▀
	─────────▐▓▓▓▓▓▓▌▐▓▓▓▓▓▓▌
	──────────▐▓▓▓▓▌──▐▓▓▓▓▌
	─────────▄████▀────▀████▄
	─────────▀▀▀▀────────▀▀▀▀
	credits:
	http://textart4u.blogspot.com/2013/08/minions-emoticons-text-art-for-facebook.html
	https://gist.githubusercontent.com/belbomemo/b5e7dad10fa567a5fe8a/raw/4ed0c8a82a8d1b836e2de16a597afca714a36606/gistfile1.txt
	`)
}

func printBuckets(s3client  minio.Client) {
	buckets, err := s3client.ListBuckets()
	if err != nil {
		log.Printf("error listing buckets: %v\n", err)
	}
	for _, bucket := range buckets {
		log.Printf("found bucket: %v\n", bucket.Name)
	}
}

func upsertBucket(s3Client minio.Client, region string, bucket string) {
	exists, err := s3Client.BucketExists(bucket)
	if err != nil {
		log.Printf("error checking bucket exists: %v\n", err)
	}

	if !exists {
		err = s3Client.MakeBucket(bucket, region)
		if err != nil {
			log.Printf("error creating bucket %s: %v\n", bucket, err)
		}
		log.Printf("created bucket: %s", bucket)
	}
}

func addNotification(s3Client minio.Client, region string, bucket string) {
	// on bucket notification
	queueArn := minio.NewArn("minio", "sqs", region, "1", "nats")
	queueConfig := minio.NewNotificationConfig(queueArn)
	queueConfig.AddEvents(minio.ObjectCreatedAll)
	queueConfig.AddEvents(minio.ObjectRemovedAll)

	bucketNotification := minio.BucketNotification{}
	bucketNotification.AddQueue(queueConfig)
	err := s3Client.SetBucketNotification(bucket, bucketNotification)

	if err != nil {
		log.Printf("Unable to set bucket notification %v\n", err)
	} else {
		log.Printf("added bucket notification: %v", queueArn)
	}
}

func getClient(endpoint string, accessKey string, secret string, encrypt bool) (*minio.Client) {
	client, err := minio.New(endpoint, accessKey, secret, encrypt)
	if err != nil {
		log.Fatalf("unable to connect to %s: %v\n", endpoint, err)
	}
	log.Printf("connected to: %s\n", endpoint)
	return client
}

func main() {
	printMinion()

	// assumes:
	// 	1. Running minio server.
	//	2. Running natsd server.
	// open connection to remote s3 bucket.
	bucket := "minio-nats-example"
	region := "us-east-1"

	// create remote client
	remoteS3AccessKeyId := os.Getenv("MINIO_AWS_ACCESS_KEY_ID")
	remoteS3SecretKey := os.Getenv("MINIO_AWS_SECRET_ACCESS_KEY")

	s3RemoteClient := getClient("s3.amazonaws.com", remoteS3AccessKeyId, remoteS3SecretKey, true)

	s3LocalClient := getClient("10.0.1.17:9000", "9W2392IOEBUAZH6PLCHA",
		"g7XPw7gWdVoazRLoVfg1ZcV3SrcUSz3gwhUonTVC", false)

	// create the bucket in both locations, if it doesn't exist
	upsertBucket(*s3LocalClient, region, bucket)
	upsertBucket(*s3RemoteClient, region, bucket)

	// add the notification interest
	addNotification(*s3LocalClient, region, bucket)

	natsConnection, _ := nats.Connect("nats://localhost:4222")
	log.Println("Connected to NATS")

	// Subscribe to subject
	log.Print("Subscribing to subject 'bucketevents'\n")
	natsConnection.Subscribe("bucketevents", func(msg *nats.Msg) {

		// Handle the message
		notification := minio.NotificationInfo{}

		err := json.Unmarshal(msg.Data, &notification)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		}

		for _, record := range notification.Records {
			key, err := url.QueryUnescape(record.S3.Object.Key)
			if err != nil {
				log.Printf("unable to escape key name (%s): %v", record.S3.Object.Key,
					err)
			}
			switch record.EventName {
			case "s3:ObjectCreated:Put":
				localFileName := fmt.Sprintf("/tmp/%s", key)
				log.Printf("syncing object: %s/%s\n", record.S3.Bucket.Name,
					key)
				err = s3LocalClient.FGetObject(record.S3.Bucket.Name, key,
					localFileName)
				if err != nil {
					fmt.Printf("error saving file: %v\n", err)
				}

				_, err = s3RemoteClient.FPutObject(bucket, key, localFileName,
					"application/octet-stream")
				if err != nil {
					fmt.Printf("error: %v\n", err)
				}
			case "s3:ObjectRemoved:Delete":
				err = s3RemoteClient.RemoveObject(bucket, key)

				if err != nil {
					log.Printf("error deleting object: %v\n", err)
				} else {
					log.Printf("deleted object: %s\n", key)
				}
			}
		}

		// get the object from the local client

	})

	// Keep the connection alive
	runtime.Goexit()

	// transmit object to remote s3.

}