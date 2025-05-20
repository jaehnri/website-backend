package ideas

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/jaehnri/website-backend/pkg/ideas"
)

const (
	BucketEnv = "GCS_BUCKET_ENV"
	ObjectEnv = "GCS_OBJECT_ENV"
)

type IdeasGCSClient struct {
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

	obj := client.Bucket(bucketName).Object(objectName)
	return &IdeasGCSClient{
		gcsClient: client,
		object:    obj,
	}
}

// GetIdeas fetches all the ideas from a single GCS object.
func (i *IdeasGCSClient) GetIdeas(req *ideas.GetIdeasRequest) (*ideas.GetIdeasResponse, error) {
	rc, err := i.object.NewReader(context.Background())
	if err != nil {
		log.Printf("failed to create object reader for %s/%s: %v", i.object.BucketName(), i.object.ObjectName(), err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		log.Printf("failed to read object contents: %v", err)
	}

	// No need to sort since they are always appended to GCS object end.
	// TODO: What if they're appended a the top?
	return &ideas.GetIdeasResponse{
		Ideas: parseIdeas(req, string(data)),
	}, nil
}

func (i *IdeasGCSClient) PostIdea(req *ideas.PostIdeaRequest) (*ideas.PostIdeaResponse, error) {
	// 1. Read the existing content of the object
	rc, err := i.object.NewReader(context.Background())
	if err != nil {
		// Handle the case where the object might not exist yet.
		// If it's a new file, we'll just start with the new line.
		if err == storage.ErrObjectNotExist {
			log.Printf("Object gs://%s/%s does not exist. Creating it with the new line.", i.object.BucketName(), i.object.ObjectName())
			// No existing content, so currentContent will be empty.
		} else {
			return nil, fmt.Errorf("failed to create object reader for %s/%s: %v", i.object.BucketName(), i.object.ObjectName(), err)
		}
	}
	defer func() {
		if rc != nil { // Ensure rc is not nil before closing
			rc.Close()
		}
	}()

	var currentContent string
	if rc != nil { // Only read if the object existed
		dataBytes, err := io.ReadAll(rc)
		if err != nil {
			return nil, fmt.Errorf("failed to read object contents: %v", err)
		}
		currentContent = string(dataBytes)
	}

	// 2. Modify the content in memory by appending the new line
	// Ensure a newline if content already exists and doesn't end with one.
	if currentContent != "" && !strings.HasSuffix(currentContent, "\n") {
		currentContent += "\n"
	}

	// TODO: Encode the Idea into a proto?
	idea := idea(req)
	jsonIdea, err := json.Marshal(idea)
	if err != nil {
		return nil, fmt.Errorf("failed to get json idea: %v", err)
	}
	newContent := currentContent + string(jsonIdea) + "\n" // Add the new line and ensure it ends with a newline

	// 3. Overwrite the original object with the new content
	// Use a writer to upload the new content. This will overwrite the existing object.
	wc := i.object.NewWriter(context.Background())
	wc.ContentType = "text/plain"

	if _, err := fmt.Fprint(wc, newContent); err != nil {
		wc.Close() // Close writer even on write error
		return nil, fmt.Errorf("failed to write new content to object: %v", err)
	}

	if err := wc.Close(); err != nil {
		log.Printf("failed to close object writer: %v", err)
		return nil, err
	}

	return &ideas.PostIdeaResponse{
		Idea: idea,
	}, nil
}

// Since every idea is a single line, parseIdeas receives all ideas in a single
// string and parses line-by-line.
func parseIdeas(req *ideas.GetIdeasRequest, fileContent string) []*ideas.Idea {
	gcsLines := strings.Split(fileContent, "\n")
	limit := min(req.Limit, len(gcsLines))

	// TODO: Use req.Offset too
	ideasList := make([]*ideas.Idea, 0, limit)
	for _, v := range gcsLines {
		ideasList = append(ideasList, &ideas.Idea{
			Idea: v,
		})
	}

	// TODO: Prioritize fresh ideas first
	return ideasList
}

func idea(req *ideas.PostIdeaRequest) *ideas.Idea {
	return &ideas.Idea{
		Idea: req.Idea,
		Time: time.Now(),
	}
}
