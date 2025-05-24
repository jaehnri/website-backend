package ideas

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jaehnri/website-backend/pkg/ideas"
)

const (
	DefaultOffset = 0
	DefaultLimit  = 50
)

type IdeasRepository interface {
	GetIdeas(req *ideas.GetIdeasRequest) (*ideas.GetIdeasResponse, error)
	PostIdea(req *ideas.PostIdeaRequest) (*ideas.PostIdeaResponse, error)
}

type IdeasClient struct {
	ideasRepo IdeasRepository
}

func NewIdeasClient() *IdeasClient {
	return &IdeasClient{
		ideasRepo: NewIdeasGCSClient(),
	}
}

func (s *IdeasClient) HandleIdeas(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.HandleGetIdeas(w, r)
	case http.MethodPost:
		s.HandlePostIdeas(w, r)
	default:
		// Respond with 405 Method Not Allowed for other methods
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}

}

func (s *IdeasClient) HandleGetIdeas(w http.ResponseWriter, r *http.Request) {
	ideas, err := s.ideasRepo.GetIdeas(parseGetIdeasRequest(r))
	if err != nil {
		http.Error(w, "failed to fetch my ideas", http.StatusInternalServerError)
		return
	}

	// TODO: Think about this later!
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	err = encoder.Encode(ideas)
	if err != nil {
		http.Error(w, "failed to encode json response", http.StatusInternalServerError)
		return
	}
}

func parseGetIdeasRequest(r *http.Request) *ideas.GetIdeasRequest {
	query := r.URL.Query()

	offsetStr := query.Get("offset")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = DefaultOffset
	}

	limitStr := query.Get("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = DefaultLimit
	}

	return &ideas.GetIdeasRequest{
		Offset: offset,
		Limit:  limit,
	}
}

func (s *IdeasClient) HandlePostIdeas(w http.ResponseWriter, r *http.Request) {
	req, err := parsePostIdeaRequest(r)
	if err != nil {
		http.Error(w, "failed to parse idea", http.StatusBadRequest)
		return
	}

	idea, err := s.ideasRepo.PostIdea(req)
	if err != nil {
		http.Error(w, "failed to post new idea", http.StatusInternalServerError)
		return
	}

	// TODO: Think about this later!
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	err = encoder.Encode(idea)
	if err != nil {
		http.Error(w, "failed to encode json response", http.StatusInternalServerError)
		return
	}
}

func parsePostIdeaRequest(r *http.Request) (*ideas.PostIdeaRequest, error) {
	var req ideas.PostIdeaRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}

	return &req, nil
}
