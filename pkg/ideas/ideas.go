package ideas

import "time"

// GetIdeasRequest is the HTTP request response for GET /ideas.
// Ideas can be searched using Limit and Offset.
type GetIdeasRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// GetIdeasResponse is the HTTP response for GET /ideas.
// Ideas are sortered from newest to oldest.
type GetIdeasResponse struct {
	Ideas []*Idea `json:"ideas"`
}

// PostIdeaResponse is the HTTP request for POST /ideas.
type PostIdeaRequest struct {
	Idea Idea `json:"idea"`
}

// PostIdeaResponse is the HTTP response for POST /ideas.
type PostIdeaResponse struct {
	Idea *Idea `json:"idea"`
}

// Idea represents an idea that I had. Or a thought. Or something.
type Idea struct {
	Time time.Time `json:"time"`
	Idea string    `json:"idea"`
}
