package main

func (s *Server) routes() {
	s.Router.HandleFunc("/presbyterians/", s.PresbyteriansHandler())
}
