package httpx

// ExampleURL describes another way to query an endpoint.
type ExampleURL struct {
	URL     string `json:"url"`
	Purpose string `json:"purpose"`
}

// Endpoint describes an API endpoint and provides sample requests.
type Endpoint struct {
	Name     string       `json:"name"`
	URL      string       `json:"path"`
	Examples []ExampleURL `json:"examples,omitempty"`
}
