package ideas

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
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

	return &ideas.GetIdeasResponse{
		Ideas: parseIdeas(req, data),
	}, nil
}

func (i *IdeasGCSClient) PostIdea(req *ideas.PostIdeaRequest) (*ideas.PostIdeaResponse, error) {
	// 1. Read the existing content of the object
	rc, err := i.object.NewReader(context.Background())
	if err != nil {
		// Handle the case where the object might not exist yet.
		// If it's a new file, we'll just start with the new idea.
		if err == storage.ErrObjectNotExist {
			log.Printf("Object gs://%s/%s does not exist. Creating...", i.object.BucketName(), i.object.ObjectName())
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

	var currentContent []*ideas.Idea
	if rc != nil { // Only read if the object existed
		dataBytes, err := io.ReadAll(rc)
		if err != nil {
			return nil, fmt.Errorf("failed to read object contents: %v", err)
		}

		err = json.Unmarshal(dataBytes, &currentContent)
		if err != nil {
			return nil, fmt.Errorf("failed to parse into idea array: %v", err)
		}
	}

	// 2. Modify the content in memory by appending the new idea
	// TODO: Encode the Idea into a proto? I feel really dirty using JSONs.
	idea := idea(req)

	// Note that by appending fresh ideas to the start, it's easy to use limits and offsets later
	// without sorting.
	newContent := append([]*ideas.Idea{idea}, currentContent...)
	jsonNewContent, err := json.Marshal(newContent)
	if err != nil {
		return nil, fmt.Errorf("failed to get json idea array: %v", err)
	}

	// 3. Overwrite the original object with the new content
	wc := i.object.NewWriter(context.Background())
	wc.ContentType = "application/json"

	if _, err := fmt.Fprint(wc, string(jsonNewContent)); err != nil {
		wc.Close()
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
func parseIdeas(req *ideas.GetIdeasRequest, fileContent []byte) []*ideas.Idea {
	var ideas []*ideas.Idea

	err := json.Unmarshal(fileContent, &ideas)
	if err != nil {
		log.Print("failed to unmarshal idea from gcs: ", err)
	}

	ideas = ideas[req.Offset:]
	limit := min(req.Limit, len(ideas))
	return ideas[:limit]
}

func idea(req *ideas.PostIdeaRequest) *ideas.Idea {
	return &ideas.Idea{
		Idea: req.Idea,
		Time: time.Now(),
	}
}
