# Minio NATS Demo
Use Minio to notify of changes through NATS and sync changes between two clouds (or a laptop and a cloud).

# Overview
Minio makes it easy to manage an object store with an S3 interface across multiple different platforms, 
from your local desktop to other clouds beyond AWS.

This demonstration will show you how to run a Minio object store on a local laptop, 
configure a local NATS message bus and finally replicate objects to other clouds.

![Diagram](/readme_img/diag.png?raw=true "Diagram")

# Tutorial

1. Install and run gnatsd
    ```
    go get github.com/nats-io/gnatsd; gnatsd -D -V
    ```
1. Install minio
    ```
    go get github.com/minio/minio
    ```
1. Configure minio for local NATS event subscription
    
    edit `~/.minio/config.json`
    
    set `"nats"."1"."enable": true`
    
    ``` 
    ...
    "nats": {
        "1": {
            "enable": true,
            "address": "0.0.0.0:4222",
            "subject": "bucketevents",
    ...
    ```
1. Run minio
    
    Set ~/minio-tmp/ to any directory you want to store your objects in.
    
    ```
    minio server ~/minio-tmp/
    ```
1. Run minioNATS
    ```
    go run minionats/main.go -remote s3://accessKeyId:accessSecretKey@host:port -local s3://accessKeyId:accessSecretKey@host:port
    ```
    
1. Open Browsers to your test bucket [Minio Browser](http://localhost:9000/minio/minio-nats-example/) and 
an [S3 Browser](https://console.aws.amazon.com/s3/buckets/minio-nats-example)

1. Upload a File to your Minio Browser. Watch it automatically get added to your S3 browser

1. Delete a File from your Minio Browser. Watch it automatically get removed from your S3 Browser

# Usage
```
Usage of demo-minio-nats:
  -bucket string
    	bucket to test with (default "minio-nats-example")
  -local string
    	local S3 URL in format s3://accessKeyId:accessSecretKey@host:port
  -nats string
    	NATS URL in format nats://user:password@host:port (default "nats://localhost:4222")
  -region string
    	region to create and maintain bucket (default "us-east-1")
  -remote string
    	remote S3 URL in format s3://accessKeyId:accessSecretKey@host:port
  -tmpDir string
    	temporary directory for copying files (default "/tmp/")
```

# Additional Reading
[Publish Minio Events via NATS](https://blog.minio.io/part-4-5-publish-minio-events-via-nats-79114ea5cd29#.s2sifywij)

[NATS Blog](http://nats.io/blog/)
