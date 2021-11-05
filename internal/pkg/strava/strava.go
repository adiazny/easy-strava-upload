package strava

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const (
	stravaTokenURL     = "https://www.strava.com/oauth/token"
	activitiesEndpoint = "https://www.strava.com/api/v3/activities"
)

// Config Struct
type Config struct {
	StravaClientID     string
	StravaClientSecret string
	StravaRefreshToken string
	Scopes             []string
}

type provider struct {
	providerName string
	config       *Config
	httpClient   *http.Client
}

type stravaRefreshResponse struct {
	TokenType    string `json:"token_type"`
	AccesToken   string `json:"access_token"`
	ExpiresAt    int    `json:"expires_at"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

func PostActivity(uiRequest *http.Request, c *Config) error {
	sp := newProvider("strava", c)

	client := &http.Client{}

	//Parse Form params coming in from UI
	uiRequest.ParseForm()
	urlValues := uiRequest.Form

	req, err := http.NewRequest("POST", activitiesEndpoint, strings.NewReader(urlValues.Encode()))
	if err != nil {
		log.Printf("Error creating HTTP request %s: %v\n", activitiesEndpoint, err)
		return err
	}

	access, _, _ := sp.RefreshToken(c.StravaRefreshToken)
	bearer := fmt.Sprintf("Bearer %s", access)
	req.Header.Add("Authorization", bearer)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making HTTP POST request to Strava /activities: %v\n", err)
		return err

	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		log.Printf("HTTP /activites reponse code: %v\n", resp.StatusCode)
		log.Printf("Strava /activities response: %v\n", resp)
		return err
	}

	return nil
}

func newProvider(name string, c *Config) *provider {
	p := new(provider)
	p.providerName = name
	p.config = c
	p.httpClient = http.DefaultClient
	return p
}

// RefreshToken refreshes access token
func (p *provider) RefreshToken(rt string) (access, refresh string, err error) {
	var tokenURL string
	var formData url.Values

	tokenURL = stravaTokenURL
	formData = url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {p.config.StravaClientID},
		"client_secret": {p.config.StravaClientSecret},
		"refresh_token": {rt},
	}

	encodedFormData := formData.Encode()

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(encodedFormData))
	if err != nil {
		log.Printf("Error creating HTTP Request %s: %v\n", tokenURL, err)
		return access, refresh, err
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		log.Printf("Error making HTTP POST to Strava OAuth /token: %v\n", err)
		return access, refresh, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var stravaRefreshResp stravaRefreshResponse
	json.Unmarshal(b, &stravaRefreshResp)
	access = stravaRefreshResp.AccesToken
	refresh = stravaRefreshResp.RefreshToken

	return

}
