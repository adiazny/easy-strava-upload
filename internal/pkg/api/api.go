package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/adiazny/easy-strava-upload/internal/pkg/strava"
	"github.com/sirupsen/logrus"
)

type Server struct {
	Log            *logrus.Entry
	Router         *http.ServeMux
	StravaProvider *strava.Provider
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}

func (s *Server) handleAboutEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("A simple app to easily and quikly upload a manual Strava activity to a Strava Profile\n"))
	}
}

func (server *Server) handleActivities() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case "GET":
			fmt.Fprintln(w, "Get Request Successful")
		case "POST":
			err := server.StravaProvider.PostActivity(req)
			if err != nil {
				log.Printf("Error Posting to Strava API: %v", err)
			}

		default:
			fmt.Fprintf(w, "Only GET, HEAD and POST allowed")
		}
	}
}
