package ideas

import (
	"context"
	"io"
	"log"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/jaehnri/website-backend/pkg/ideas"
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
func (i *IdeasGCSClient) GetIdeas(req *ideas.GetIdeasRequest) (*ideas.GetIdeasResponse, error) {
	rc, err := i.object.NewReader(context.Background())
	if err != nil {
		log.Printf("failed to create object reader for %s/%s: %v", i.bucket.BucketName(), i.object.ObjectName(), err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		log.Printf("failed to read object contents: %v", err)
	}

	// No need to sort since they are always appended to GCS object end.
	return &ideas.GetIdeasResponse{
		Ideas: parseIdeas(string(data)),
	}, nil
}

func (i *IdeasGCSClient) PostIdea(req *ideas.PostIdeaRequest) (*ideas.PostIdeaResponse, error) {
	// 	log.Printf("Attempting to append \"%s\" to gs://%s/%s", idea, i.bucket.BucketName(), i.object.ObjectName())

	// 	// 1. Read the existing content of the object
	// 	rc, err := i.object.NewReader(context.Background())
	// 	if err != nil {
	// 		// Handle the case where the object might not exist yet.
	// 		// If it's a new file, we'll just start with the new line.
	// 		if err == storage.ErrObjectNotExist {
	// 			log.Printf("Object gs://%s/%s does not exist. Creating it with the new line.", i.bucket.BucketName(), i.object.ObjectName())
	// 			// No existing content, so currentContent will be empty.
	// 		} else {
	// 			return fmt.Errorf("failed to create object reader for %s/%s: %v", i.bucket.BucketName(), i.object.ObjectName(), err)
	// 		}
	// 	}
	// 	defer func() {
	// 		if rc != nil { // Ensure rc is not nil before closing
	// 			rc.Close()
	// 		}
	// 	}()

	// 	var currentContent string
	// 	if rc != nil { // Only read if the object existed
	// 		dataBytes, err := io.ReadAll(rc)
	// 		if err != nil {
	// 			return fmt.Errorf("failed to read object contents: %v", err)
	// 		}
	// 		currentContent = string(dataBytes)
	// 	}

	// 	// 2. Modify the content in memory by appending the new line
	// 	// Ensure a newline if content already exists and doesn't end with one.
	// 	if currentContent != "" && !strings.HasSuffix(currentContent, "\n") {
	// 		currentContent += "\n"
	// 	}
	// 	newContent := currentContent + idea + "\n" // Add the new line and ensure it ends with a newline

	// 	// 3. Overwrite the original object with the new content
	// 	// Use a writer to upload the new content. This will overwrite the existing object.
	// 	wc := i.object.NewWriter(context.Background())
	// 	// Optional: Set content type if known, e.g., "text/plain"
	// 	wc.ContentType = "text/plain"

	// 	if _, err := fmt.Fprint(wc, newContent); err != nil {
	// 		wc.Close() // Close writer even on write error
	// 		return fmt.Errorf("failed to write new content to object: %v", err)
	// 	}

	// 	if err := wc.Close(); err != nil {
	// 		log.Printf("failed to close object writer: %v", err)
	// 		return err
	// 	}

	// 	log.Printf("Successfully appended line to gs://%s/%s", i.bucket.BucketName(), i.object.ObjectName())
	return nil, nil
}

// Since every idea is a single line, parseIdeas receives all ideas in a single
// string and parses line-by-line.
func parseIdeas(req *ideas.GetIdeasRequest, fileContent string) []*ideas.Idea {
	gcsLines := strings.Split(fileContent, "\n")

	limit := len(gcsLines)
	if req.Limit < limit {
		limit = req.Limit
	}

	ideasList := make([]*ideas.Idea, 0, limit)
	for _, v := range gcsLines {
		ideasList = append(ideasList, &ideas.Idea{
			Idea: v,
		})
	}

	// TODO: I need to prioritize fresh ideas first.
	return ideasList
}
