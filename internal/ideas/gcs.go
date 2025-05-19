package ideas

import (
	"context"
	"io"
	"log"
	"os"
	"strings"

	"cloud.google.com/go/storage"
)

const (
	BucketEnv = "GCS_BUCKET_ENV"
	ObjectEnv = "GCS_OBJECT_ENV"
)

type IdeasGCSClient struct {
	bucket    *storage.BucketHandle
	object    *storage.ObjectHandle
	gcsClient *storage.Client
}

func NewIdeasGCSClient() *IdeasGCSClient {
	bucketName, exists := os.LookupEnv(BucketEnv)
	if !exists {
		log.Panic("couldn't retrieve GCS bucket")
	}

	objectName, exists := os.LookupEnv(ObjectEnv)
	if !exists {
		log.Panic("couldn't retrieve GCS object")
	}

	client, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("Failed to create GCS client: %v", err)
	}

	bucket := client.Bucket(bucketName)
	obj := bucket.Object(objectName)
	return &IdeasGCSClient{
		gcsClient: client,

		bucket: bucket,
		object: obj,
	}
}

// GetIdeas fetches all the ideas from a single GCS object.
func (i *IdeasGCSClient) GetIdeas() ([]string, error) {
	rc, err := i.object.NewReader(context.Background())
	if err != nil {
		log.Printf("failed to create object reader for %s/%s: %v", i.bucket.BucketName(), i.object.ObjectName(), err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		log.Printf("failed to read object contents: %v", err)
	}

	return parseIdeas(string(data)), nil
}

func (i *IdeasGCSClient) PostIdea(idea string) error {
	return nil
}

// Since every idea is a single line, parseIdeas receives all ideas in a single
// string and parses line-by-line.
func parseIdeas(fileContent string) []string {
	return strings.Split(fileContent, "\n")
}
