package api

import (
	"fmt"
	"net/http"
)

type Server struct {
	router *http.ServeMux
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) handleAboutEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("A simple app to easily and quikly upload a manual Strava activity to a Strava Profile\n"))
	}
}

func (s *Server) handleActivities() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			fmt.Fprintln(w, "Get Request Successful")
		case "POST":
			// unmarshall request body into strava input body
			// stravaInput := strava.getBody(r.Body)
			// POST to strava API
		default:
			fmt.Fprintf(w, "Only GET, HEAD and POST allowed")
		}
	}
}

// NewServer returns a new Server value
func NewServer() *Server {
	s := &Server{
		router: http.NewServeMux(),
	}
	s.routes()

	return s
}
