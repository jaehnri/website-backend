package ideas

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"cloud.google.com/go/storage"
)

const (
	bucketName = "joaohenri-website-test-bucket"
	objectName = "test.txt"
)

type IdeasGCSClient struct {
	gcsClient *storage.Client
}

func NewIdeasGCSClient() *IdeasGCSClient {
	client, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("Failed to create GCS client: %v", err)
	}
	return &IdeasGCSClient{
		gcsClient: client,
	}
}

func (i *IdeasGCSClient) GetIdeas() ([]string, error) {
	bucket := i.gcsClient.Bucket(bucketName)
	obj := bucket.Object(objectName)

	rc, err := obj.NewReader(context.Background())
	if err != nil {
		log.Printf("failed to create object reader for %s/%s: %v", bucketName, objectName, err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		log.Printf("failed to read object contents: %v", err)
	}

	fileContent := string(data)
	fmt.Printf("contents of gs://%s/%s:\n%s\n", bucketName, objectName, string(data))

	return parseIdeas(fileContent), nil
}

func (i *IdeasGCSClient) PostIdea(idea string) error {
	return nil
}

// Since every idea is a single line, parseIdeas receives all ideas in a single
// string and parses line-by-line.
func parseIdeas(fileContent string) []string {
	return strings.Split(fileContent, "\n")
}
