package ideas

import "time"

// GetIdeasRequest is the HTTP request response for GET /ideas.
// Ideas can be filtered using Limit and Offset.
type GetIdeasRequest struct {
	// How many ideas to fetch. Defaults to 50.
	Limit int `json:"limit"`

	// GET /ideas always fetches newest ideas. Increase offset
	// to fetch older ones.
	Offset int `json:"offset"`
}

// GetIdeasResponse is the HTTP response for GET /ideas.
// Ideas are sortered from newest to oldest.
type GetIdeasResponse struct {
	Ideas []*Idea `json:"ideas"`
}

// PostIdeaResponse is the HTTP request for POST /ideas.
type PostIdeaRequest struct {
	Idea string `json:"idea"`
}

// PostIdeaResponse is the HTTP response for POST /ideas.
type PostIdeaResponse struct {
	*Idea
}

// Idea represents an idea that I had. Or a thought. Or something.
type Idea struct {
	Time time.Time `json:"time"`
	Idea string    `json:"idea"`
}
