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
	stravaTokenURL = "https://www.strava.com/oauth/token"
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

	activitiesEndpoint := "https://www.strava.com/api/v3/activities"

	client := &http.Client{}

	//Parse Form params coming in from UI
	uiRequest.ParseForm()
	urlValues := uiRequest.Form

	req, err := http.NewRequest("POST", activitiesEndpoint, strings.NewReader(urlValues.Encode()))
	if err != nil {
		log.Printf("NewRequest Log Err: %v\n", err)
		return fmt.Errorf("Error: %v", err)
	}

	access, _, _ := sp.RefreshToken(c.StravaRefreshToken)
	bearer := fmt.Sprintf("Bearer %s", access)
	req.Header.Add("Authorization", bearer)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("client.Do Log Err: %v\n", err)
		return fmt.Errorf("Error: %v", err)

	}
	defer resp.Body.Close()

	log.Printf("Strava POST Response Status: %v", resp.Status)
	log.Printf("Strava Response: %v", resp)
	if resp.StatusCode != 201 {
		return fmt.Errorf("Strava Error: %v", resp.Status)
	}

	return err
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
	log.Printf("Provider: %s Inside RefreshToken()", p.providerName)

	var tokenURL string
	var formData url.Values

	switch p.providerName {
	case "strava":
		tokenURL = stravaTokenURL
		formData = url.Values{
			"grant_type":    {"refresh_token"},
			"client_id":     {p.config.StravaClientID},
			"client_secret": {p.config.StravaClientSecret},
			"refresh_token": {rt},
		}
	}

	encodedFormData := formData.Encode()

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(encodedFormData))
	if err != nil {
		log.Printf("NewRequest Log Err: %v\n", err)
		return access, refresh, fmt.Errorf("Error: %v", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		log.Printf("client.Do Log Err: %v\n", err)
		return access, refresh, fmt.Errorf("Error: %v", err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	switch p.providerName {
	case "strava":
		var stravaRefreshResp stravaRefreshResponse
		json.Unmarshal(b, &stravaRefreshResp)
		log.Printf("RefreshResponse: %+v\n", stravaRefreshResp)
		access = stravaRefreshResp.AccesToken
		refresh = stravaRefreshResp.RefreshToken
		return
	}
	return

}
