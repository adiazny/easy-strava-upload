package strava

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	stravaTokenURL     = "https://www.strava.com/oauth/token"
	activitiesEndpoint = "https://www.strava.com/api/v3/activities"
)

type Provider struct {
	Log          *logrus.Entry
	ProviderName string
	Config       *Config
	HTTPClient   *http.Client
}

// Config Struct
type Config struct {
	StravaClientID     string
	StravaClientSecret string
	StravaRefreshToken string
	Scopes             []string
}
type stravaRefreshResponse struct {
	TokenType    string `json:"token_type"`
	AccesToken   string `json:"access_token"`
	ExpiresAt    int    `json:"expires_at"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

func (provider *Provider) PostActivity(uiRequest *http.Request) error {

	client := &http.Client{}

	//Parse Form params coming in from UI
	uiRequest.ParseForm()
	urlValues := uiRequest.Form

	req, err := http.NewRequest("POST", activitiesEndpoint, strings.NewReader(urlValues.Encode()))
	if err != nil {
		provider.Log.Infof("Error creating HTTP request %s: %v\n", activitiesEndpoint, err)
		return err
	}

	access, _, _ := provider.RefreshToken(provider.Config.StravaRefreshToken)
	bearer := fmt.Sprintf("Bearer %s", access)
	req.Header.Add("Authorization", bearer)

	resp, err := client.Do(req)
	if err != nil {
		provider.Log.Infof("Error making HTTP POST request to Strava /activities: %v\n", err)
		return err

	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		provider.Log.Infof("HTTP /activites reponse code: %v\n", resp.StatusCode)
		provider.Log.Infof("Strava /activities response: %v\n", resp)
		return err
	}

	return nil
}

// RefreshToken refreshes the OAUTH access token
func (provider *Provider) RefreshToken(rt string) (access, refresh string, err error) {
	var tokenURL string
	var formData url.Values

	tokenURL = stravaTokenURL
	formData = url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {provider.Config.StravaClientID},
		"client_secret": {provider.Config.StravaClientSecret},
		"refresh_token": {rt},
	}

	encodedFormData := formData.Encode()

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(encodedFormData))
	if err != nil {
		provider.Log.Infof("Error creating HTTP Request %s: %v\n", tokenURL, err)
		return access, refresh, err
	}

	resp, err := provider.HTTPClient.Do(req)
	if err != nil {
		provider.Log.Infof("Error making HTTP POST to Strava OAuth /token: %v\n", err)
		return access, refresh, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		provider.Log.Fatal(err)
	}

	var stravaRefreshResp stravaRefreshResponse
	json.Unmarshal(b, &stravaRefreshResp)
	access = stravaRefreshResp.AccesToken
	refresh = stravaRefreshResp.RefreshToken

	return
}

func NewProvider(log *logrus.Entry, name string, c *Config) *Provider {
	p := new(Provider)
	p.Log = log
	p.ProviderName = name
	p.Config = c
	p.HTTPClient = http.DefaultClient
	return p
}
