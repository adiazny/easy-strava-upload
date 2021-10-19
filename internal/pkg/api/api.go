package api

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/adiazny/easy-strava-upload/internal/pkg/strava"
)

type Server struct {
	router *http.ServeMux
	config *strava.Config
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
			// POST to strava API
			err := strava.PostActivity(r, s.config)
			if err != nil {
				log.Printf("Error Posting to Strava API: %v", err)
			}

		default:
			fmt.Fprintf(w, "Only GET, HEAD and POST allowed")
		}
	}
}

// NewServer returns a new Server value
func NewServer() *Server {
	s := &Server{
		router: http.NewServeMux(),
		config: loadConfig(),
	}
	s.routes()

	log.Println("Easy-Strava-Upload Application Started")
	return s
}

func loadConfig() *strava.Config {

	log.Println("Loading Strava configuration.")
	return &strava.Config{
		StravaClientID:     getEnvVar("STRAVA_CLIENT_ID"),
		StravaClientSecret: getEnvVar("STRAVA_CLIENT_SECRET"),
		StravaRefreshToken: getEnvVar("STRAVA_REFRESH_TOKEN"),
	}
}

func getEnvVar(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		log.Fatalf("Environment variable %v is not set", key)
	}
	if val == "" {
		log.Fatalf("Environment variable %v is empty", key)
	}
	return val
}
