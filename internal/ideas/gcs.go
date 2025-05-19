package ideas

type IdeasGCSClient struct {
	Some bool
}

func NewIdeasGCSClient() *IdeasGCSClient {
	return &IdeasGCSClient{
		Some: true,
	}
}

func (i *IdeasGCSClient) GetIdeas() ([]string, error) {
	return []string{"123", "1234"}, nil
}

func (i *IdeasGCSClient) PostIdea(idea string) error {
	return nil
}
