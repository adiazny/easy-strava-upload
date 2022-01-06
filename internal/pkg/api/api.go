package api

import (
	"fmt"
	"net/http"
	"text/template"

	"github.com/adiazny/easy-strava-upload/internal/pkg/strava"
	"github.com/markbates/goth/gothic"
	"github.com/sirupsen/logrus"
)

type Server struct {
	Log            *logrus.Entry
	Router         *http.ServeMux
	StravaProvider *strava.Provider
	ProviderIndex  *ProviderIndex
}

type ProviderIndex struct {
	Providers    []string
	ProvidersMap map[string]string
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
				server.Log.Infof("Error Posting to Strava API: %w", err)
			}

		default:
			fmt.Fprintf(w, "Only GET, HEAD and POST allowed")
		}
	}
}

func (s *Server) handleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t, _ := template.New("foo").Parse(indexTemplate)
		t.Execute(w, s.ProviderIndex)
	}
}

func (s *Server) handleAuthentication() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.RawQuery
		fmt.Println("Query", query)
		// try to get the user without re-authenticating
		if gothUser, err := gothic.CompleteUserAuth(w, r); err == nil {
			t, _ := template.New("foo").Parse(userTemplate)
			t.Execute(w, gothUser)
		} else {
			if err != nil {
				fmt.Println("Get Store Error", err.Error())
			}
			gothic.BeginAuthHandler(w, r)
		}
	}
}

func (s *Server) handleCallback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		t, _ := template.New("foo").Parse(userTemplate)
		t.Execute(w, user)
	}
}

func (s *Server) handleLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gothic.Logout(w, r)
		w.Header().Set("Location", "/")
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

var indexTemplate = `{{range $key,$value:=.Providers}}
    <p><a href="/auth/{{$value}}?provider={{$value}}">Log in with {{index $.ProvidersMap $value}}</a></p>
{{end}}`

var userTemplate = `
<p><a href="/logout/{{.Provider}}">logout</a></p>
<p>Name: {{.Name}} [{{.LastName}}, {{.FirstName}}]</p>
<p>Email: {{.Email}}</p>
<p>NickName: {{.NickName}}</p>
<p>Location: {{.Location}}</p>
<p>AvatarURL: {{.AvatarURL}} <img src="{{.AvatarURL}}"></p>
<p>Description: {{.Description}}</p>
<p>UserID: {{.UserID}}</p>
<p>AccessToken: {{.AccessToken}}</p>
<p>ExpiresAt: {{.ExpiresAt}}</p>
<p>RefreshToken: {{.RefreshToken}}</p>
`
