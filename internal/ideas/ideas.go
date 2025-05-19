package ideas

import (
	"encoding/json"
	"net/http"
)

type IdeasRepository interface {
	GetIdeas() ([]string, error)
	PostIdea(string) error
}

type IdeasClient struct {
	ideasRepo IdeasRepository
}

func NewIdeasClient() *IdeasClient {
	return &IdeasClient{
		ideasRepo: NewIdeasGCSClient(),
	}
}

func (s *IdeasClient) HandleGetIdeas(w http.ResponseWriter, r *http.Request) {
	ideas, err := s.ideasRepo.GetIdeas()
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
