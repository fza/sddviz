// Package graph answers the front end's questions about a decision graph.
package graph

// Service is the surface the front end calls. It holds no state until a
// repository is opened.
type Service struct{}

func NewService() *Service {
	return &Service{}
}

// BackendVersion reports the running back end, which is how the front end
// confirms the seam between the two modules works.
func (s *Service) BackendVersion() string {
	return version
}

const version = "0.1.0"
