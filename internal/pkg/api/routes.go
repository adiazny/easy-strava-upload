package api

func (s *Server) Routes() {
	s.Router.HandleFunc("/", s.handleHome())
	s.Router.HandleFunc("/auth/strava", s.handleAuthentication())
	s.Router.HandleFunc("/auth/strava/callback", s.handleCallback())
	s.Router.HandleFunc("/auth/logout/strava", s.handleLogout())
	s.Router.HandleFunc("/about", s.handleAboutEndpoint())
	s.Router.HandleFunc("/activities", s.handleActivities())

}
