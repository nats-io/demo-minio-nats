# Minio NATS Demo
Use Minio to notify of changes through NATS and sync changes between two clouds (or a laptop and a cloud).

# Overview
Minio makes it easy to manage an object store with an S3 interface across multiple different platforms, from your local desktop to other clouds beyond AWS.

This demonstration will show you how to run a Minio object store on a local laptop, configure a local NATS message bus and finally replicate objects to other clouds.

# Future Enhancements
Configure a load balancer (cloudflare? backbone?) to serve from either the S3 endpoint or your local laptop (other cloud) based on availability.
