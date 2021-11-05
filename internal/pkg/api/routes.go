package api

func (s *Server) Routes() {
	s.Router.HandleFunc("/about", s.handleAboutEndpoint())
	s.Router.HandleFunc("/activities", s.handleActivities())
}
